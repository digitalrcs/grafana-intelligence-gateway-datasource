package plugin

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAcquireRejectsFifthCallWithoutQueueing(t *testing.T) {
	g := testGateway(t, func(http.ResponseWriter, *http.Request) { t.Error("provider must not be contacted") })
	t.Cleanup(func() {
		for len(g.concurrent) > 0 {
			g.release()
		}
	})
	for index := 0; index < 4; index++ {
		if err := g.acquire(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	started := time.Now()
	err := g.acquire(ctx)
	if statusFor(err) != http.StatusTooManyRequests || err.Error() != "data source is busy; retry later" || time.Since(started) >= 500*time.Millisecond || len(g.concurrent) != 4 {
		t.Fatalf("fifth request was queued or changed active slots: %v", err)
	}
	g.release()
	if err := g.acquire(context.Background()); err != nil || len(g.concurrent) != 4 {
		t.Fatalf("released slot was not reusable: %v", err)
	}
}

func TestAcquireMapsCancelledContextWithoutTakingSlot(t *testing.T) {
	g := testGateway(t, func(http.ResponseWriter, *http.Request) { t.Error("provider must not be contacted") })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := g.acquire(ctx); statusFor(err) != http.StatusRequestTimeout || len(g.concurrent) != 0 {
		t.Fatalf("cancelled request acquired a slot or was not mapped safely: %v", err)
	}
}

type unreadRequestBody struct{ read bool }

func (r *unreadRequestBody) Read([]byte) (int, error) { r.read = true; return 0, io.EOF }
func (*unreadRequestBody) Close() error               { return nil }

func TestBusyResourceRejectsBeforeReadingImageBody(t *testing.T) {
	g := testGateway(t, func(http.ResponseWriter, *http.Request) { t.Error("provider must not be contacted") })
	for index := 0; index < 4; index++ {
		if err := g.acquire(context.Background()); err != nil {
			t.Fatal(err)
		}
		defer g.release()
	}
	body := &unreadRequestBody{}
	request := httptest.NewRequest(http.MethodPost, "/chat/completions", nil)
	request.Body = body
	response := httptest.NewRecorder()
	(&Datasource{gateway: g}).handleChat(response, request)
	assertPrivateJSONResponse(t, response)
	if response.Code != http.StatusTooManyRequests || body.read || !strings.Contains(response.Body.String(), "retry later") {
		t.Fatalf("busy request was read or incorrectly rejected: %d %s", response.Code, response.Body.String())
	}
}

func TestRequestSlotsStayHeldUntilResponseBodyCloses(t *testing.T) {
	for _, resource := range []string{"chat", "models"} {
		t.Run(resource, func(t *testing.T) {
			finish := make(chan struct{})
			var calls atomic.Int32
			g := testGateway(t, func(w http.ResponseWriter, req *http.Request) {
				calls.Add(1)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.(http.Flusher).Flush()
				// Headers have arrived, but a slow provider is still generating its
				// response. This must consume a slot until its body is closed.
				select {
				case <-finish:
				case <-req.Context().Done():
				}
			})
			t.Cleanup(func() { close(finish) })
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			request := func() (*http.Response, error) {
				if resource == "models" {
					return g.requestModels(ctx)
				}
				return g.requestChat(ctx, gatewayRequest{Prompt: &prompt{User: "assess"}})
			}
			var responses []*http.Response
			for index := 0; index < 4; index++ {
				response, err := request()
				if err != nil {
					t.Fatal(err)
				}
				responses = append(responses, response)
				t.Cleanup(func() { closeBody(response.Body) })
			}
			if len(g.concurrent) != 4 {
				t.Fatalf("response headers released active slots: %d", len(g.concurrent))
			}
			response, err := request()
			if response != nil || statusFor(err) != http.StatusTooManyRequests || calls.Load() != 4 {
				t.Fatalf("fifth request reached the slow provider: response=%v error=%v calls=%d", response, err, calls.Load())
			}
			closeBody(responses[0].Body)
			closeBody(responses[0].Body)
			if len(g.concurrent) != 3 {
				t.Fatalf("closing twice released more than one slot: %d", len(g.concurrent))
			}
			response, err = request()
			if err != nil || len(g.concurrent) != 4 || calls.Load() != 5 {
				t.Fatalf("closed body did not make one slot reusable: %v", err)
			}
			t.Cleanup(func() { closeBody(response.Body) })
		})
	}
}

func TestRequestSlotsReleaseOnProviderOrValidationErrors(t *testing.T) {
	g := testGateway(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	for _, request := range []func() (*http.Response, error){
		func() (*http.Response, error) {
			return g.requestChat(context.Background(), gatewayRequest{Prompt: &prompt{User: "assess"}})
		},
		func() (*http.Response, error) {
			return g.requestChat(context.Background(), gatewayRequest{Model: "not-allowed", Prompt: &prompt{User: "assess"}})
		},
		func() (*http.Response, error) { return g.requestModels(context.Background()) },
	} {
		response, err := request()
		if err == nil || response != nil || len(g.concurrent) != 0 {
			t.Fatalf("failed request retained a slot: response=%v error=%v active=%d", response, err, len(g.concurrent))
		}
	}
}

func TestResponseReadFailureReleasesSlotWhenClosed(t *testing.T) {
	g := testGateway(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "100")
		_, _ = io.WriteString(w, "{\"truncated\":")
	})
	response, err := g.requestModels(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeBody(response.Body) })
	if _, err := readBounded(response.Body); err == nil {
		t.Fatal("expected a truncated body read failure")
	}
	closeBody(response.Body)
	if len(g.concurrent) != 0 {
		t.Fatal("failed body read retained its slot after close")
	}
}

type countedCloseBody struct{ closes atomic.Int32 }

func (*countedCloseBody) Read([]byte) (int, error) { return 0, io.EOF }
func (body *countedCloseBody) Close() error {
	body.closes.Add(1)
	return io.ErrClosedPipe
}

func TestSlotResponseBodyCloseIsConcurrentAndIdempotent(t *testing.T) {
	underlying := &countedCloseBody{}
	var releases atomic.Int32
	body := &slotResponseBody{ReadCloser: underlying, release: func() { releases.Add(1) }}
	var callers sync.WaitGroup
	for index := 0; index < 8; index++ {
		callers.Add(1)
		go func() {
			defer callers.Done()
			if err := body.Close(); err != io.ErrClosedPipe {
				t.Errorf("close did not preserve the underlying error: %v", err)
			}
		}()
	}
	callers.Wait()
	if underlying.closes.Load() != 1 || releases.Load() != 1 {
		t.Fatalf("body close/release must run once: closes=%d releases=%d", underlying.closes.Load(), releases.Load())
	}
}

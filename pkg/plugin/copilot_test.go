package plugin

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func testCopilotGateway(t *testing.T, handler http.HandlerFunc) *gateway {
	t.Helper()
	g := testGateway(t, handler)
	g.settings.Provider = "copilot-studio"
	g.settings.TimeoutSeconds = 2
	g.settings.Secrets.DirectLineSecret = "channel-secret"
	g.baseURL.Path = "/v3/directline"
	return g
}

func writeCopilotJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func copilotTestContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func copilotReply(id, text, hint string) map[string]any {
	return map[string]any{"type": "message", "id": id, "text": text, "inputHint": hint, "replyToId": "sent-1", "from": map[string]string{"id": "agent-1", "role": "bot"}, "conversation": map[string]string{"id": "conversation-1"}}
}

func TestCopilotAssessmentUsesScopedTokenAndMultipartBytes(t *testing.T) {
	for _, withImages := range []bool{false, true} {
		t.Run(fmt.Sprint(withImages), func(t *testing.T) {
			attachments := []imageAttachment{}
			if withImages {
				attachments = []imageAttachment{testImage(t, "png", 2, 3), testImage(t, "jpeg", 3, 2)}
			}
			var userID string
			var calls atomic.Int32
			g := testCopilotGateway(t, func(w http.ResponseWriter, req *http.Request) {
				calls.Add(1)
				if req.URL.Path == "/v3/directline/tokens/generate" {
					if req.Header.Get("Authorization") != "Bearer channel-secret" {
						t.Error("channel secret missing")
					}
					var tokenRequest struct {
						User copilotActor `json:"user"`
					}
					if err := json.NewDecoder(req.Body).Decode(&tokenRequest); err != nil {
						t.Error(err)
					}
					userID = tokenRequest.User.ID
					if !strings.HasPrefix(userID, "dl_") || len(userID) != 35 {
						t.Error("missing unguessable channel identity")
					}
					writeCopilotJSON(w, copilotConversation{ID: "conversation-1", Token: "scoped-token", ExpiresIn: 1800})
					return
				}
				if req.Header.Get("Authorization") != "Bearer scoped-token" {
					t.Error("scoped conversation token not used")
				}
				switch req.URL.Path {
				case "/v3/directline/conversations":
					writeCopilotJSON(w, map[string]any{"conversationId": "conversation-1", "streamUrl": "https://untrusted.invalid/never-follow"})
				case "/v3/directline/conversations/conversation-1/activities", "/v3/directline/conversations/conversation-1/upload":
					if req.Method == http.MethodGet {
						writeCopilotJSON(w, map[string]any{"activities": []any{copilotReply("answer-1", "The chart rises.", "acceptingInput")}, "watermark": "w1"})
						return
					}
					var activity copilotActivity
					if withImages {
						if req.URL.Path != "/v3/directline/conversations/conversation-1/upload" || req.URL.Query().Get("userId") != userID {
							t.Error("incorrect upload route/user")
						}
						mediaType, parameters, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
						if err != nil || mediaType != "multipart/form-data" {
							t.Fatalf("bad multipart content type: %v", err)
						}
						reader := multipart.NewReader(req.Body, parameters["boundary"])
						count := 0
						for {
							part, err := reader.NextPart()
							if err == io.EOF {
								break
							}
							if err != nil {
								t.Fatal(err)
							}
							body, _ := io.ReadAll(part)
							if part.Header.Get("Content-Type") == "application/vnd.microsoft.activity" {
								if err := json.Unmarshal(body, &activity); err != nil {
									t.Error(err)
								}
							} else {
								if count >= len(attachments) {
									t.Fatal("too many image parts")
								}
								want, _ := base64.StdEncoding.DecodeString(strings.SplitN(attachments[count].DataURL, ",", 2)[1])
								if !bytes.Equal(body, want) || !strings.HasPrefix(part.FileName(), "image-") {
									t.Error("uploaded bytes or safe filename differ")
								}
								count++
							}
						}
						if count != 2 {
							t.Error("images missing")
						}
					} else if err := json.NewDecoder(req.Body).Decode(&activity); err != nil {
						t.Error(err)
					}
					if activity.From.ID != userID || !strings.Contains(activity.Text, "look at data") || !strings.Contains(activity.Text, "Be careful") {
						t.Error("request context lost")
					}
					if strings.Contains(activity.Text, "channel-secret") || strings.Contains(activity.Text, "max_tokens") {
						t.Error("secrets or unsupported controls in transcript")
					}
					writeCopilotJSON(w, map[string]string{"id": "sent-1"})
				default:
					t.Errorf("unexpected request %s", req.URL.Path)
					http.Error(w, "unexpected", 500)
				}
			})
			enableTestVision(g)
			response, err := g.requestChat(context.Background(), gatewayRequest{Prompt: &prompt{System: "Be careful", User: "look at data"}, Images: attachments})
			if err != nil {
				t.Fatal(err)
			}
			defer closeBody(response.Body)
			body, _ := io.ReadAll(response.Body)
			answer, err := extractAnswer(body)
			if err != nil || answer != "The chart rises." || calls.Load() != 4 {
				t.Fatalf("unexpected answer/calls: %s %v %d", answer, err, calls.Load())
			}
			for _, secret := range []string{"channel-secret", "scoped-token", "conversation-1", "untrusted.invalid"} {
				if bytes.Contains(body, []byte(secret)) {
					t.Error("provider internals leaked")
				}
			}
		})
	}
}

func TestCopilotHealthOnlyExchangesToken(t *testing.T) {
	var calls atomic.Int32
	g := testCopilotGateway(t, func(w http.ResponseWriter, req *http.Request) {
		calls.Add(1)
		if req.URL.Path != "/v3/directline/tokens/generate" || req.Method != http.MethodPost {
			t.Error("health contacted agent")
		}
		writeCopilotJSON(w, copilotConversation{ID: "conversation-1", Token: "scoped-token", ExpiresIn: 1800})
	})
	if err := g.checkCopilotHealth(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatal("unexpected network calls")
	}
}

func TestCopilotPollingCorrelatesDeduplicatesAndPreservesWatermarks(t *testing.T) {
	var calls atomic.Int32
	g := testCopilotGateway(t, func(w http.ResponseWriter, req *http.Request) {
		switch calls.Add(1) {
		case 1:
			unrelated := copilotReply("unrelated", "wrong assessment", "acceptingInput")
			unrelated["replyToId"] = "other-turn"
			user := copilotReply("echo", "user echo", "acceptingInput")
			user["from"] = map[string]string{"id": "user-1", "role": "user"}
			foreign := copilotReply("foreign", "wrong conversation", "acceptingInput")
			foreign["conversation"] = map[string]string{"id": "other-conversation"}
			writeCopilotJSON(w, map[string]any{"activities": []any{unrelated, user, foreign, copilotReply("progress", "Checking the image.", "")}, "watermark": "opaque+ /value"})
		case 2:
			if req.URL.Query().Get("watermark") != "opaque+ /value" {
				t.Error("watermark changed")
			}
			writeCopilotJSON(w, map[string]any{"activities": []any{copilotReply("progress", "Checking the image.", "")}, "watermark": nil})
		default:
			if req.URL.Query().Get("watermark") != "opaque+ /value" {
				t.Error("nil watermark replaced previous")
			}
			writeCopilotJSON(w, map[string]any{"activities": []any{copilotReply("answer", "Final assessment.", "acceptingInput")}})
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, err := g.copilotPoll(ctx, "default-model", copilotConversation{ID: "conversation-1", Token: "scoped-token"}, "user-1", "sent-1", time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer closeBody(response.Body)
	body, _ := io.ReadAll(response.Body)
	answer, _ := extractAnswer(body)
	if answer != "Checking the image.\n\nFinal assessment." || calls.Load() != 3 {
		t.Fatalf("incorrect correlated assessment %q", answer)
	}
}

func TestCopilotRejectsInteractiveAuthWithoutFollowingURLs(t *testing.T) {
	for _, cardType := range []string{"application/vnd.microsoft.card.oauth", "application/vnd.microsoft.card.signin"} {
		t.Run(cardType, func(t *testing.T) {
			var calls atomic.Int32
			g := testCopilotGateway(t, func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				activity := copilotReply("login", "Click to log in", "acceptingInput")
				delete(activity, "replyToId")
				activity["attachments"] = []any{map[string]any{"contentType": cardType, "content": map[string]any{"buttons": []any{map[string]string{"type": "signin", "value": "https://private.invalid/login"}}}}}
				writeCopilotJSON(w, map[string]any{"activities": []any{activity}})
			})
			_, err := g.copilotPoll(copilotTestContext(t), "default-model", copilotConversation{ID: "conversation-1", Token: "scoped-token"}, "user-1", "sent-1", time.Millisecond)
			if err == nil || !strings.Contains(err.Error(), "interactive user sign-in/SSO") || strings.Contains(err.Error(), "private.invalid") || calls.Load() != 1 {
				t.Fatalf("expected safe auth error: %v", err)
			}
		})
	}
	for _, activity := range []copilotActivity{{Name: "signin/tokenExchange"}, {SuggestedActions: json.RawMessage(`{"actions":[{"type":"signin","value":"https://never-follow.invalid"}]}`)}} {
		if !copilotRequiresSignIn(activity) {
			t.Error("sign-in action not detected")
		}
	}
}

func TestCopilotDoesNotTreatProgressOrUnrelatedTextAsCompletion(t *testing.T) {
	for _, mode := range []string{"progress", "unrelated", "empty", "nontext", "cancelled"} {
		t.Run(mode, func(t *testing.T) {
			g := testCopilotGateway(t, func(w http.ResponseWriter, _ *http.Request) {
				activity := copilotReply("answer", "Still working", "")
				if mode == "unrelated" {
					activity["replyToId"] = "other"
					activity["inputHint"] = "acceptingInput"
				}
				if mode == "empty" || mode == "nontext" {
					activity["text"] = ""
				}
				if mode == "empty" {
					activity["inputHint"] = "acceptingInput"
				}
				writeCopilotJSON(w, map[string]any{"activities": []any{activity}})
			})
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			if mode == "cancelled" {
				cancel()
			}
			_, err := g.copilotPoll(ctx, "default-model", copilotConversation{ID: "conversation-1", Token: "scoped-token"}, "user-1", "sent-1", time.Millisecond)
			if err == nil {
				t.Fatal("incomplete assessment accepted")
			}
			switch mode {
			case "empty", "nontext":
				if statusFor(err) != http.StatusBadGateway {
					t.Errorf("wrong empty/nontext error: %v", err)
				}
			case "cancelled":
				if statusFor(err) != http.StatusRequestTimeout {
					t.Errorf("wrong cancelled error: %v", err)
				}
			default:
				if statusFor(err) != http.StatusGatewayTimeout {
					t.Errorf("wrong incomplete error: %v", err)
				}
			}
		})
	}
}

func TestCopilotChannelErrorsAreBoundedAndSanitized(t *testing.T) {
	for _, mode := range []string{"credentials", "rate", "redirect", "large", "invalidjson", "wrongtype", "path", "lifetime"} {
		t.Run(mode, func(t *testing.T) {
			var calls atomic.Int32
			g := testCopilotGateway(t, func(w http.ResponseWriter, req *http.Request) {
				calls.Add(1)
				switch mode {
				case "credentials":
					http.Error(w, "credentialsecret", http.StatusForbidden)
				case "rate":
					http.Error(w, "credentialsecret", http.StatusTooManyRequests)
				case "redirect":
					http.Redirect(w, req, "/credentialsecret", http.StatusTemporaryRedirect)
				case "large":
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(strings.Repeat("x", copilotControlBytes+1)))
				case "invalidjson":
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte("credentialsecret"))
				case "wrongtype":
					w.Header().Set("Content-Type", "text/html")
					_, _ = w.Write([]byte("credentialsecret"))
				case "path":
					writeCopilotJSON(w, copilotConversation{ID: "../credentialsecret", Token: "secret", ExpiresIn: 100})
				case "lifetime":
					writeCopilotJSON(w, copilotConversation{ID: "conversation-1", Token: "secret", ExpiresIn: 0})
				}
			})
			_, err := g.copilotToken(context.Background(), "")
			if err == nil || strings.Contains(err.Error(), "credentialsecret") || calls.Load() != 1 {
				t.Fatalf("unsafe error/redirect: %v", err)
			}
		})
	}
}

func TestCopilotOverallTimeoutAndMissingSecret(t *testing.T) {
	var calls atomic.Int32
	g := testCopilotGateway(t, func(w http.ResponseWriter, req *http.Request) {
		calls.Add(1)
		_, _ = io.Copy(io.Discard, req.Body)
		<-req.Context().Done()
	})
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancel()
	_, err := g.requestChat(ctx, gatewayRequest{Prompt: &prompt{User: "look"}})
	if statusFor(err) != http.StatusGatewayTimeout {
		t.Fatalf("expected bounded deadline: %v", err)
	}
	g.settings.Secrets.DirectLineSecret = ""
	_, err = g.requestChat(context.Background(), gatewayRequest{Prompt: &prompt{User: "look"}})
	if err == nil || !strings.Contains(err.Error(), "channel secret") || calls.Load() != 1 {
		t.Fatalf("secret fallback allowed: %v", err)
	}
}

func TestCopilotNoPublicImageURLsOrMultipartHeaderInjection(t *testing.T) {
	_, _, err := copilotPayload([]chatMessage{{Role: "user", Content: "look"}}, []imageAttachment{{Name: "remote", DataURL: "https://internal.invalid/image.png"}}, "dl_user")
	if err == nil {
		t.Fatal("remote image accepted")
	}
	image := testImage(t, "png", 2, 2)
	image.Name = "local-path\r\nInjected: yes"
	body, contentType, err := copilotPayload([]chatMessage{{Role: "user", Content: "look"}}, []imageAttachment{image}, "dl_user")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(body, []byte("Injected")) || !strings.HasPrefix(contentType, "multipart/form-data") {
		t.Fatal("image filename leaked/injected")
	}
}

func TestCopilotWatermarkCannotChangeEndpoint(t *testing.T) {
	g := testCopilotGateway(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v3/directline/conversations/conversation-1/activities" || req.URL.Query().Get("watermark") != "https://private.invalid/?x=secret" {
			t.Error("query was treated as an endpoint")
		}
		writeCopilotJSON(w, map[string]any{"activities": []any{}})
	})
	var result any
	_, err := g.copilotJSON(context.Background(), http.MethodGet, "conversations/conversation-1/activities", url.Values{"watermark": []string{"https://private.invalid/?x=secret"}}, "scoped-token", "", nil, &result, 1000)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCopilotEachAssessmentUsesANewIdentityAndConversation(t *testing.T) {
	identities := make(map[string]bool)
	issued := 0
	g := testCopilotGateway(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/v3/directline/tokens/generate" {
			var request struct {
				User copilotActor `json:"user"`
			}
			if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
				t.Error(err)
			}
			if identities[request.User.ID] {
				t.Error("channel identity reused")
			}
			identities[request.User.ID] = true
			issued++
			writeCopilotJSON(w, copilotConversation{ID: fmt.Sprintf("conversation-%d", issued), Token: fmt.Sprintf("token-%d", issued), ExpiresIn: 1800})
			return
		}
		if req.Header.Get("Authorization") != fmt.Sprintf("Bearer token-%d", issued) {
			t.Error("token crossed assessments")
		}
		if req.URL.Path == "/v3/directline/conversations" {
			writeCopilotJSON(w, map[string]string{"conversationId": fmt.Sprintf("conversation-%d", issued)})
			return
		}
		if req.URL.Path != fmt.Sprintf("/v3/directline/conversations/conversation-%d/activities", issued) {
			t.Error("conversation crossed assessments")
		}
		if req.Method == http.MethodPost {
			writeCopilotJSON(w, map[string]string{"id": "sent-1"})
			return
		}
		answer := copilotReply("answer-1", "Assessment", "acceptingInput")
		answer["conversation"] = map[string]string{"id": fmt.Sprintf("conversation-%d", issued)}
		writeCopilotJSON(w, map[string]any{"activities": []any{answer}})
	})
	for range 2 {
		response, err := g.requestChat(context.Background(), gatewayRequest{Prompt: &prompt{User: "look"}})
		if err != nil {
			t.Fatal(err)
		}
		closeBody(response.Body)
	}
	if len(identities) != 2 || issued != 2 {
		t.Fatal("assessments were not isolated")
	}
}

func TestCopilotEndOfConversationAndActivityLimits(t *testing.T) {
	for _, mode := range []string{"end", "too-many", "watermark", "invalid-id"} {
		t.Run(mode, func(t *testing.T) {
			g := testCopilotGateway(t, func(w http.ResponseWriter, _ *http.Request) {
				activities := []any{copilotReply("answer-1", "Assessment", "")}
				end := copilotReply("end", "", "")
				end["type"] = "endOfConversation"
				activities = append(activities, end)
				watermark := "w1"
				switch mode {
				case "too-many":
					activities = nil
					for index := 0; index <= copilotMaxActivities; index++ {
						activities = append(activities, copilotReply(fmt.Sprint(index), "progress", ""))
					}
				case "watermark":
					watermark = strings.Repeat("w", 2049)
				case "invalid-id":
					activities = []any{copilotReply("", "bad", "acceptingInput")}
				}
				writeCopilotJSON(w, map[string]any{"activities": activities, "watermark": watermark})
			})
			response, err := g.copilotPoll(copilotTestContext(t), "default-model", copilotConversation{ID: "conversation-1", Token: "token"}, "user-1", "sent-1", time.Millisecond)
			if mode == "end" {
				if err != nil {
					t.Fatal(err)
				}
				closeBody(response.Body)
			} else if statusFor(err) != http.StatusBadGateway {
				t.Fatalf("expected bounded protocol failure: %v", err)
			}
		})
	}
}

func TestCopilotRequestPolicyStillAppliesBeforeChannelContact(t *testing.T) {
	var calls atomic.Int32
	g := testCopilotGateway(t, func(w http.ResponseWriter, _ *http.Request) { calls.Add(1); http.Error(w, "unexpected", 500) })
	attachment := testImage(t, "png", 2, 2)
	for _, request := range []gatewayRequest{
		{Prompt: &prompt{User: "look"}, Model: "unapproved-agent"},
		{Prompt: &prompt{User: "look"}, Images: []imageAttachment{attachment}},
		{Prompt: &prompt{User: "look"}, Stream: true},
		{Prompt: &prompt{User: strings.Repeat("a", maxRequestBytes)}},
	} {
		if _, err := g.requestChat(context.Background(), request); err == nil {
			t.Error("invalid policy request accepted")
		}
	}
	if calls.Load() != 0 {
		t.Fatal("policy rejection contacted channel")
	}
}

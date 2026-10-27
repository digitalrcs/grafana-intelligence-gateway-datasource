package main

import "testing"

func TestTransportAcceptsImageEnvelopeWithBoundedHeadroom(t *testing.T) {
	// Do not silently fall back to the SDK's 16 MiB limit: valid image bodies
	// may be up to 18 MiB before protobuf and data-source metadata are added.
	if got := manageOptions().GRPCSettings.MaxReceiveMsgSize; got != 20<<20 {
		t.Fatalf("unexpected gRPC receive bound: %d", got)
	}
}

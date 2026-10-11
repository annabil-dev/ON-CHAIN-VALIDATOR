package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConsoleProgressRPC(t *testing.T) {
	catching := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/status":
			_, _ = fmt.Fprintf(w, `{"result":{"sync_info":{"latest_block_height":"3822","catching_up":%t}}}`, catching)
		case "/net_info":
			_, _ = fmt.Fprint(w, `{"result":{"n_peers":"1"}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	c := server.Client()
	got, err := getNodeStatus(context.Background(), c, server.URL)
	if err != nil || got != "SYNCING | height 3822" {
		t.Fatalf("unexpected progress %q: %v", got, err)
	}
	peers := getPeerCount(context.Background(), c, server.URL)
	if peers != "1" {
		t.Fatalf("unexpected peers %q", peers)
	}
	catching = false
	got, err = getNodeStatus(context.Background(), c, server.URL)
	if err != nil || got != "SYNCED | height 3822" {
		t.Fatalf("unexpected synced message %q: %v", got, err)
	}
}

func TestConsoleStatusBadRPC(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	_, err := getNodeStatus(context.Background(), srv.Client(), srv.URL)
	if err == nil {
		t.Fatal("Expected error on failed RPC")
	}
}

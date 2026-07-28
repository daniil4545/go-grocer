package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// withTestServer points fnsURL/httpClient at srv and restores the real values
// once the test finishes, so tests never touch the network.
func withTestServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	origURL, origClient := fnsURL, httpClient
	fnsURL, httpClient = srv.URL, srv.Client()
	t.Cleanup(func() { fnsURL, httpClient = origURL, origClient })
}

// withShortRetryDelays shrinks the code=2/4 retry delays so exhaustion tests
// don't take tens of seconds, and restores them afterwards.
func withShortRetryDelays(t *testing.T, delay time.Duration) {
	t.Helper()

	origCode2, origCode4 := retryDelayCode2, retryDelayCode4
	retryDelayCode2, retryDelayCode4 = delay, delay
	t.Cleanup(func() { retryDelayCode2, retryDelayCode4 = origCode2, origCode4 })
}

func writeCheckResponse(t *testing.T, w http.ResponseWriter, code int, data any) {
	t.Helper()

	rawData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	body, err := json.Marshal(CheckResponse{Code: code, Data: rawData})
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	if _, err := w.Write(body); err != nil {
		t.Fatalf("write response: %v", err)
	}
}

func TestGetReceipt_Success(t *testing.T) {
	var requests int32

	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		writeCheckResponse(t, w, 1, CheckData{JSON: Receipt{
			TotalSum: 12345,
			User:     "ВкусВилл",
			Items:    []Item{{Name: "Молоко", Sum: 8990}},
		}})
	})

	receipt, err := GetReceipt(context.Background(), "token", []byte("photo"))
	if err != nil {
		t.Fatalf("get receipt: %v", err)
	}
	if receipt.User != "ВкусВилл" || receipt.TotalSum != 12345 {
		t.Fatalf("receipt = %+v, want User=ВкусВилл TotalSum=12345", receipt)
	}
	if got := atomic.LoadInt32(&requests); got != 1 {
		t.Fatalf("requests = %d, want 1", got)
	}
}

func TestGetReceipt_ExhaustsRetries(t *testing.T) {
	withShortRetryDelays(t, time.Millisecond)

	tests := []struct {
		name string
		code int
	}{
		{name: "code 2", code: 2},
		{name: "code 4", code: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requests int32

			withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&requests, 1)
				writeCheckResponse(t, w, tt.code, Receipt{})
			})

			_, err := GetReceipt(context.Background(), "token", []byte("photo"))
			if err == nil {
				t.Fatalf("err = nil, want max retry attempts error")
			}
			if got := atomic.LoadInt32(&requests); got != 5 {
				t.Fatalf("requests = %d, want 5", got)
			}
		})
	}
}

func TestGetReceipt_CancelDuringSleep(t *testing.T) {
	withShortRetryDelays(t, 500*time.Millisecond)

	var requests int32
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		writeCheckResponse(t, w, 2, Receipt{})
	})

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)

	_, err := GetReceipt(ctx, "token", []byte("photo"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if got := atomic.LoadInt32(&requests); got != 1 {
		t.Fatalf("requests = %d, want 1", got)
	}
}

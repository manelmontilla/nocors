package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestDisablesCORS(t *testing.T) {
	upstreamHandler := func(w http.ResponseWriter, r *http.Request) {
		// Check we don't receive preflight requests upstream.
		if w.Header().Get("Access-Control-Request-Method") != "" {
			t.Errorf("upstream test server received a preflight request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
	upstream := httptest.NewServer(http.HandlerFunc(upstreamHandler))
	defer upstream.Close()

	u, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatalf("unexpected error parsing test server url %+v: %+v", u, err)
	}
	nocors := httptest.NewServer(newNoCorsReverseProxy(u.Host))
	defer nocors.Close()

	// Check that the proxy intercepts and generates proper responses for
	// preflight requests.
	req, err := http.NewRequest(http.MethodOptions, nocors.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error creating a test request: %+v", err)
	}

	req.Header.Add("Access-Control-Request-Method", http.MethodGet)
	req.Header.Add("Access-Control-Request-Headers", "NO-CORS")
	wantOrigin := "origin"
	req.Header.Add("Origin", wantOrigin)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("error sending test request: %v", err)
	}

	gotOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	if gotOrigin != wantOrigin {
		t.Errorf("got unexpected Access-Control-Allow-Origin header, got: %s, want: %s", gotOrigin, wantOrigin)
	}

	gotAllowMethods := resp.Header.Get("Access-Control-Allow-Methods")
	if gotAllowMethods != http.MethodGet {
		t.Errorf("got unexpected Access-Control-Allow-Methods header, got: %s, want: %s", gotAllowMethods, http.MethodGet)
	}

	gotAllowAccessControl := resp.Header.Get("Access-Control-Allow-Headers")
	if gotAllowAccessControl != "NO-CORS" {
		t.Errorf("got unexpected Access-Control-Allow-Headers header, got: %s, want: %s", gotAllowAccessControl, "NO-CORS")
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("got unexpected http status, got: %s, want: %s", resp.Status, http.StatusText(http.StatusNoContent))
	}
}

func TestUpstream(t *testing.T) {
	var (
		wantStatus  = http.StatusOK
		wantPayload = []byte("ok")
	)
	upstreamHandler := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(wantPayload); err != nil {
			t.Fatalf("error writing test payload: %+v", err)
		}
		w.WriteHeader(wantStatus)
	}
	upstream := httptest.NewServer(http.HandlerFunc(upstreamHandler))
	defer upstream.Close()
	u, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatalf("unexpected error parsing test server url %+v: %+v", u, err)
	}

	nocors := httptest.NewServer(newNoCorsReverseProxy(u.Host))
	defer nocors.Close()

	// Check that the proxy sends no preflight request upstream.
	req, err := http.NewRequest(http.MethodGet, nocors.URL+"/", nil)
	if err != nil {
		t.Fatalf("unexpected error creating a test request: %+v", err)
	}

	wantOrigin := "origin"
	req.Header.Add("Origin", wantOrigin)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error sending test request: %v", err)
	}
	gotOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	if gotOrigin != wantOrigin {
		t.Errorf("got unexpected Access-Control-Allow-Origin header, got: %s, want: %s", gotOrigin, wantOrigin)
	}
	if resp.StatusCode != wantStatus {
		t.Errorf("got unexpected response status code, got: %d, want: %d", resp.StatusCode, wantStatus)
	}

	gotPayload, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("unexpected error reading response payload: %v", err)
	}

	if string(gotPayload) != string(wantPayload) {
		t.Errorf("got unexpected response payload, got: %s, want: %s", string(gotPayload), string(wantPayload))
	}
}

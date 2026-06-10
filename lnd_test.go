package lnc

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// newTestLnd spins up an httptest server with the provided handler and returns
// an *Lnd pointed at it.
func newTestLnd(t *testing.T, handler http.HandlerFunc) (*Lnd, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	host, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse test server url: %v", err)
	}
	host.Path = "/"
	return &Lnd{
		Host:     host,
		Client:   srv.Client(),
		Macaroon: "deadbeef",
	}, srv
}

func TestIdentityPubkey(t *testing.T) {
	const want = "02abc123"
	lnd, srv := newTestLnd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/getinfo" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Grpc-Metadata-macaroon") != "deadbeef" {
			t.Errorf("missing macaroon header")
		}
		json.NewEncoder(w).Encode(map[string]any{"identity_pubkey": want})
	})
	defer srv.Close()

	got, err := lnd.IdentityPubkey()
	if err != nil {
		t.Fatalf("IdentityPubkey: %v", err)
	}
	if got != want {
		t.Fatalf("IdentityPubkey = %q, want %q", got, want)
	}
}

func TestIdentityPubkeyEmpty(t *testing.T) {
	lnd, srv := newTestLnd(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{})
	})
	defer srv.Close()

	if _, err := lnd.IdentityPubkey(); err == nil {
		t.Fatal("expected error for empty identity_pubkey, got nil")
	}
}

func TestIdentityPubkeyError(t *testing.T) {
	lnd, srv := newTestLnd(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "permission denied", http.StatusForbidden)
	})
	defer srv.Close()

	if _, err := lnd.IdentityPubkey(); err == nil {
		t.Fatal("expected error on non-200 response, got nil")
	}
}

func TestSignMessage(t *testing.T) {
	const wantSig = "sig-zbase32"
	msg := []byte("lnproxy:v1:announce:npub")
	lnd, srv := newTestLnd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/signmessage" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Msg string `json:"msg"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		// LND REST expects base64-encoded bytes for `msg`.
		decoded, err := base64.StdEncoding.DecodeString(req.Msg)
		if err != nil {
			t.Fatalf("msg is not valid base64: %v", err)
		}
		if string(decoded) != string(msg) {
			t.Errorf("server received msg %q, want %q", decoded, msg)
		}
		json.NewEncoder(w).Encode(map[string]any{"signature": wantSig})
	})
	defer srv.Close()

	got, err := lnd.SignMessage(msg)
	if err != nil {
		t.Fatalf("SignMessage: %v", err)
	}
	if got != wantSig {
		t.Fatalf("SignMessage = %q, want %q", got, wantSig)
	}
}

func TestSignMessageEmpty(t *testing.T) {
	lnd, srv := newTestLnd(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{})
	})
	defer srv.Close()

	if _, err := lnd.SignMessage([]byte("x")); err == nil {
		t.Fatal("expected error for empty signature, got nil")
	}
}

func TestSignMessageError(t *testing.T) {
	lnd, srv := newTestLnd(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	defer srv.Close()

	if _, err := lnd.SignMessage([]byte("x")); err == nil {
		t.Fatal("expected error on non-200 response, got nil")
	}
}

// Ensure *Lnd satisfies both interfaces.
var (
	_ LN     = (*Lnd)(nil)
	_ Signer = (*Lnd)(nil)
)

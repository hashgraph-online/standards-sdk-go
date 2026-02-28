package registrybroker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestClientBasics(t *testing.T) {
	client, _ := NewClient(RegistryBrokerClientOptions{
		BaseURL: "http://example.com",
		APIKey:  "key1",
	})
	if client.BaseURL() != "http://example.com/api/v1" {
		t.Fatalf("expected url http://example.com/api/v1, got %s", client.BaseURL())
	}
	client.SetAPIKey("key2")

	// test pointer helpers
	i := 5
	b := true
	vi, ok := intPointerValue(&i)
	if !ok || vi != 5 { t.Fatal("expected 5") }
	vi2, ok2 := intPointerValue(nil)
	if ok2 || vi2 != 0 { t.Fatal("expected 0") }

	vb, okb := boolPointerValue(&b)
	if !okb || vb != true { t.Fatal("expected true") }
	vb2, okb2 := boolPointerValue(nil)
	if okb2 || vb2 != false { t.Fatal("expected false") }
	
	q := url.Values{}
	addQueryInt(q, "count", &i)
	if q.Get("count") != "5" { t.Fatal("expected 5") }
	
	addQueryInt(q, "count", nil)
	
	addQueryBool(q, "active", &b)
	if q.Get("active") != "true" { t.Fatal("expected query") }
	
	addQueryBool(q, "active", nil)
}

func TestRequestNoResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusNoContent)
}))
	defer ts.Close()

	client, _ := NewRegistryBrokerClient(RegistryBrokerClientOptions{
BaseURL: ts.URL,
})

	err := client.requestNoResponse(context.Background(), "POST", "/", nil, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	tsFail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusBadRequest)
w.Write([]byte(`{"success": false, "error": "bad"}`))
}))
	defer tsFail.Close()

	clientFail, _ := NewRegistryBrokerClient(RegistryBrokerClientOptions{
BaseURL: tsFail.URL,
})
	err2 := clientFail.requestNoResponse(context.Background(), "POST", "/", nil, nil)
	if err2 == nil {
		t.Fatal("expected error")
	}
}

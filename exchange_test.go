package grantprovider

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockExchangeFetcher implementa ExchangeFetcher para tests sin servidor HTTP.
type mockExchangeFetcher struct {
	resp ExchangeReponse
	err  error
}

func (m *mockExchangeFetcher) Execute(_ ExchangeRequest) (ExchangeReponse, error) {
	return m.resp, m.err
}

// exchangeTestServer crea un httptest.Server que responde con la ExchangeReponse dada.
func exchangeTestServer(t *testing.T, resp ExchangeReponse) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("esperaba método POST, obtuve: %s", r.Method)
		}
		if r.Header.Get("Content-Type") != ContentTypeJSON {
			t.Errorf("esperaba Content-Type %s, obtuve: %s", ContentTypeJSON, r.Header.Get("Content-Type"))
		}
		w.Header().Set("Content-Type", ContentTypeJSON)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("error codificando respuesta del servidor: %v", err)
		}
	}))
}

func TestExchangeServiceExecute_Success(t *testing.T) {
	expectedData := map[string]any{"client_id": "my-client-id", "client_secret": "my-client-secret"}
	server := exchangeTestServer(t, ExchangeReponse{
		Data:    expectedData,
		Message: "credenciales obtenidas",
	})
	defer server.Close()

	svc := ExchangeFetcherService{
		Provider:         "test-provider",
		SessionID:        "session-123",
		ExchangeEndpoint: server.URL,
	}

	resp, err := svc.Execute(ExchangeRequest{
		Operation: OperationGetClientCredentials,
		OTT:       "test-ott-token",
	})

	if err != nil {
		t.Fatalf("esperaba nil error, obtuve: %v", err)
	}
	if resp.Message != "credenciales obtenidas" {
		t.Errorf("mensaje inesperado: %q", resp.Message)
	}
	if resp.Data == nil {
		t.Error("esperaba Data no nil")
	}
}

func TestExchangeServiceExecute_BuildsCorrectURL(t *testing.T) {
	var capturedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", ContentTypeJSON)
		_ = json.NewEncoder(w).Encode(ExchangeReponse{Data: nil, Message: "ok"})
	}))
	defer server.Close()

	svc := ExchangeFetcherService{
		Provider:         "atlassian",
		SessionID:        "session-001",
		ExchangeEndpoint: server.URL,
	}

	_, err := svc.Execute(ExchangeRequest{
		Operation: OperationGetClientCredentials,
		OTT:       "ott-abc",
	})

	if err != nil {
		t.Fatalf("esperaba nil error, obtuve: %v", err)
	}
	if !strings.Contains(capturedPath, "atlassian") {
		t.Errorf("esperaba provider en la URL, path obtenido: %s", capturedPath)
	}
	if !strings.Contains(capturedPath, "session-001") {
		t.Errorf("esperaba session_id en la URL, path obtenido: %s", capturedPath)
	}
}

func TestExchangeServiceExecute_HTTPError(t *testing.T) {
	svc := ExchangeFetcherService{
		Provider:         "test-provider",
		SessionID:        "session-123",
		ExchangeEndpoint: "http://localhost:1",
	}

	_, err := svc.Execute(ExchangeRequest{
		Operation: OperationGetClientCredentials,
		OTT:       "test-ott",
	})

	if err == nil {
		t.Fatal("esperaba error por endpoint inalcanzable, obtuve nil")
	}
}

// ========== Pruebas para GetClientCredentialsService con servidor HTTP ==========

func TestGetClientCredentialsService_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", ContentTypeJSON)
		_ = json.NewEncoder(w).Encode(ExchangeReponse{
			Data:    ClientCredentialsData{ClientID: "cid", ClientSecret: "csecret"},
			Message: "ok",
		})
	}))
	defer server.Close()

	fetcher := &ExchangeFetcherService{
		Provider:         "atlassian",
		SessionID:        "session-001",
		ExchangeEndpoint: server.URL,
	}
	svc := &GetClientCredentialsService{ExchangeFetcher: fetcher, OTT: "my-ott"}

	creds, err := svc.Execute()

	if err != nil {
		t.Fatalf("esperaba nil error, obtuve: %v", err)
	}
	if creds.ClientID != "cid" {
		t.Errorf("ClientID inesperado: %q", creds.ClientID)
	}
	if creds.ClientSecret != "csecret" {
		t.Errorf("ClientSecret inesperado: %q", creds.ClientSecret)
	}
}

func TestGetClientCredentialsService_NetworkError(t *testing.T) {
	fetcher := &ExchangeFetcherService{
		Provider:         "atlassian",
		SessionID:        "session-001",
		ExchangeEndpoint: "http://localhost:1",
	}
	svc := &GetClientCredentialsService{ExchangeFetcher: fetcher, OTT: "my-ott"}

	_, err := svc.Execute()

	if err == nil {
		t.Fatal("esperaba error de red, obtuve nil")
	}
}

// ========== Pruebas para GetClientCredentialsService con mock (sin servidor HTTP) ==========

func TestGetClientCredentialsService_MockSuccess(t *testing.T) {
	mock := &mockExchangeFetcher{
		resp: ExchangeReponse{
			Data:    ClientCredentialsData{ClientID: "mock-client", ClientSecret: "mock-secret"},
			Message: "ok",
		},
	}
	svc := &GetClientCredentialsService{ExchangeFetcher: mock, OTT: "any-ott"}

	creds, err := svc.Execute()

	if err != nil {
		t.Fatalf("esperaba nil error, obtuve: %v", err)
	}
	if creds.ClientID != "mock-client" {
		t.Errorf("ClientID inesperado: %q", creds.ClientID)
	}
	if creds.ClientSecret != "mock-secret" {
		t.Errorf("ClientSecret inesperado: %q", creds.ClientSecret)
	}
}

func TestGetClientCredentialsService_MockFetcherError(t *testing.T) {
	mock := &mockExchangeFetcher{
		err: errors.New("error simulado del fetcher"),
	}
	svc := &GetClientCredentialsService{ExchangeFetcher: mock, OTT: "any-ott"}

	_, err := svc.Execute()

	if err == nil {
		t.Fatal("esperaba error del fetcher, obtuve nil")
	}
	if err.Error() != "error simulado del fetcher" {
		t.Errorf("mensaje de error inesperado: %v", err)
	}
}

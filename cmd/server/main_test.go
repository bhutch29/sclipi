package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleScpiRequestQuery(t *testing.T) {
	var err error
  config = &Config{}
	preferences = &Preferences{}
	response, status, err := postScpi("*IDN?")
	if err != nil {
		t.Errorf("postScpi failed: %v", err)
	}
	if status != http.StatusOK {
		t.Errorf("expected status OK, got %s", http.StatusText(status))
	}
	if strings.TrimSpace(response.Response) != "*IDN?" {
		t.Errorf("expected echo of *IDN? got %v", response)
	}
}

func TestHandleScpiRequestCommand(t *testing.T) {
	response, status, err := postScpi(":FREQ 1000")
	if err != nil {
		t.Errorf("postScpi failed: %v", err)
	}
	if status != http.StatusOK {
		t.Errorf("expected status OK, got %s", http.StatusText(status))
	}
	if response.Response != "" {
		t.Errorf("expected empty response got %v", response)
	}
}

func TestHandleScpiRequestMustBePost(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/scpi", strings.NewReader("*IDN?"))
	w := httptest.NewRecorder()
	handleScpiRequest(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status Method Not Allowed, got %s", res.Status)
	}
}

func postScpi(scpi string) (scpiResponse, int, error) {
	req := httptest.NewRequest(http.MethodPost, "/scpi?simulated=true", strings.NewReader(scpi))
	w := httptest.NewRecorder()
	handleScpiRequest(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return scpiResponse{}, res.StatusCode, err
	}
	var response scpiResponse
	if err := json.Unmarshal(data, &response); err != nil {
    return scpiResponse{}, res.StatusCode, err
  }
  return response, res.StatusCode, nil
}

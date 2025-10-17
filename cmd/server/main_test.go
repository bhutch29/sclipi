package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleScpiRequestQuery(t *testing.T) {
    response, status, err := postScpi(scpiRequestBody{Type: "query", Scpi: "*IDN?", Simulated: true})
    if err != nil {
        t.Errorf("postScpi failed: %v", err)
    }
    if status != http.StatusOK {
        t.Errorf("expected status OK, got %s", http.StatusText(status))
    }
    if response != "*IDN?" {
        t.Errorf("expected echo of *IDN? got %v", response)
    }
}

func TestHandleScpiRequestCommand(t *testing.T) {
    response, status, err := postScpi(scpiRequestBody{Type: "command", Scpi: "*IDN?", Simulated: true})
    if err != nil {
        t.Errorf("postScpi failed: %v", err)
    }
    if status != http.StatusOK {
        t.Errorf("expected status OK, got %s", http.StatusText(status))
    }
    if response != "" {
        t.Errorf("expected empty response got %v", response)
    }
}

func TestHandleScpiRequestMustBePost(t *testing.T) {
    body := scpiRequestBody{Type: "query", Scpi: "*IDN?", Simulated: true}
    bodyData, _ := json.Marshal(body)
    req := httptest.NewRequest(http.MethodGet, "/scpi", bytes.NewReader(bodyData))
    w := httptest.NewRecorder()
    handleScpiRequest(w, req)
    res := w.Result()
    if res.StatusCode != http.StatusBadRequest {
        t.Errorf("expected status Bad Request, got %s", res.Status)
    }
}

func postScpi(body any) (string, int, error) {
    bodyData, _ := json.Marshal(body)
    req := httptest.NewRequest(http.MethodPost, "/scpi", bytes.NewReader(bodyData))
    w := httptest.NewRecorder()
    handleScpiRequest(w, req)
    res := w.Result()
    defer res.Body.Close()
    data, err := io.ReadAll(res.Body)
    if err != nil {
        return "", 0, err
    }
    return strings.TrimSpace(string(data)), res.StatusCode, nil
}

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

func TestHandleScpiRequest(t *testing.T) {
    response, status, err := post("/scpi", scpiRequestBody{Type: "query", Scpi: "*IDN?", Simulated: true})
    if err != nil {
        t.Errorf("post failed: %v", err)
    }
    if status != http.StatusOK {
        t.Errorf("expected status OK, got %s", http.StatusText(status))
    }
    if response != "*IDN?" {
        t.Errorf("expected echo of *IDN? got %v", response)
    }
}

func post(uri string, body any) (string, int, error) {
    bodyData, err := json.Marshal(body)
    req := httptest.NewRequest(http.MethodPost, uri, bytes.NewReader(bodyData))
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

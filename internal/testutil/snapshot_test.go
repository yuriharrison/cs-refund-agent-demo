package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotTransport_RecordMode(t *testing.T) {
	dir := t.TempDir()
	origDir := snapshotDir()
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	// Override snapshot path via temp file approach
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"test","choices":[{"message":{"content":"hello"}}]}`))
	}))
	defer server.Close()

	pairs := []RequestResponsePair{}
	filePath := filepath.Join(dir, "record_test.json")

	transport := &SnapshotTransport{
		t:         t,
		filePath:  filePath,
		recording: true,
	}

	req, err := http.NewRequest(http.MethodPost, server.URL, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("round trip: %v", err)
	}
	resp.Body.Close()

	if len(transport.pairs) != 1 {
		t.Fatalf("expected 1 recorded pair, got %d", len(transport.pairs))
	}

	transport.save()

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read snapshot file: %v", err)
	}
	if err := json.Unmarshal(data, &pairs); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair in file, got %d", len(pairs))
	}

	_ = origDir
}

func TestSnapshotTransport_ReplayMode(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "replay_test.json")

	pairs := []RequestResponsePair{
		{
			Request:  json.RawMessage(`{}`),
			Response: json.RawMessage(`{"id":"replay","choices":[{"message":{"content":"replayed"}}]}`),
			Status:   200,
		},
	}
	data, err := json.Marshal(pairs)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	transport := &SnapshotTransport{
		t:        t,
		filePath: filePath,
		pairs:    pairs,
		recording: false,
	}

	req, err := http.NewRequest(http.MethodPost, "http://example.com/v1/chat/completions", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("round trip: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["id"] != "replay" {
		t.Errorf("expected replayed response, got %v", body["id"])
	}
}

func TestSnapshotTransport_SequentialMatching(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "sequential_test.json")

	pairs := []RequestResponsePair{
		{Response: json.RawMessage(`{"id":"first"}`), Status: 200},
		{Response: json.RawMessage(`{"id":"second"}`), Status: 200},
	}
	data, _ := json.Marshal(pairs)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	transport := &SnapshotTransport{
		t:         t,
		filePath:  filePath,
		pairs:     pairs,
		recording: false,
	}

	for i, wantID := range []string{"first", "second"} {
		req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
		resp, err := transport.RoundTrip(req)
		if err != nil {
			t.Fatalf("round trip %d: %v", i, err)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if body["id"] != wantID {
			t.Errorf("request %d: want id %q, got %v", i, wantID, body["id"])
		}
	}

	if transport.RequestsRemaining() != 0 {
		t.Errorf("expected 0 remaining requests, got %d", transport.RequestsRemaining())
	}
}

func TestWriteAllScenarioSnapshots(t *testing.T) {
	if err := WriteAllScenarioSnapshots(); err != nil {
		t.Fatalf("write scenario snapshots: %v", err)
	}

	for name := range ScenarioSnapshots() {
		path := filepath.Join(snapshotDir(), name+".json")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("missing snapshot file: %s", path)
		}
	}
}

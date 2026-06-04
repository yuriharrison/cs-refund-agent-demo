package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

type RequestResponsePair struct {
	Request  json.RawMessage `json:"request"`
	Response json.RawMessage `json:"response"`
	Status   int             `json:"status"`
}

type SnapshotTransport struct {
	t         *testing.T
	filePath  string
	pairs     []RequestResponsePair
	index     int
	mu        sync.Mutex
	recording bool
}

func NewSnapshotTransport(t *testing.T, name string) *SnapshotTransport {
	t.Helper()

	dir := snapshotDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed to create snapshot dir: %v", err)
	}

	filePath := filepath.Join(dir, name+".json")
	st := &SnapshotTransport{
		t:        t,
		filePath: filePath,
	}

	data, err := os.ReadFile(filePath)
	if err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &st.pairs); err != nil {
			t.Fatalf("failed to parse snapshot %s: %v", filePath, err)
		}
		st.recording = false
	} else {
		st.recording = true
		st.pairs = nil
		t.Cleanup(func() {
			st.save()
		})
	}

	return st
}

func (st *SnapshotTransport) IsRecording() bool {
	return st.recording
}

func SkipIfRecordingWithoutAPIKey(t *testing.T, transport *SnapshotTransport) {
	t.Helper()
	if transport.IsRecording() && os.Getenv("OPEN_ROUTER_API_KEY") == "" {
		t.Skip("no OPEN_ROUTER_API_KEY set, skipping recording")
	}
}

func snapshotDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("internal", "testutil", "snapshots")
	}
	return filepath.Join(filepath.Dir(file), "snapshots")
}

func (st *SnapshotTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	if st.recording {
		return st.record(req)
	}
	return st.replay(req)
}

func (st *SnapshotTransport) record(req *http.Request) (*http.Response, error) {
	var reqBody []byte
	if req.Body != nil {
		var err error
		reqBody, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("reading request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	resp.Body.Close()

	pair := RequestResponsePair{
		Request:  json.RawMessage(reqBody),
		Response: json.RawMessage(respBody),
		Status:   resp.StatusCode,
	}
	st.pairs = append(st.pairs, pair)

	resp.Body = io.NopCloser(bytes.NewReader(respBody))
	return resp, nil
}

func (st *SnapshotTransport) replay(req *http.Request) (*http.Response, error) {
	if st.index >= len(st.pairs) {
		st.t.Fatalf("snapshot %s: unexpected HTTP request %d (only %d recorded)", st.filePath, st.index+1, len(st.pairs))
	}

	pair := st.pairs[st.index]
	st.index++

	status := pair.Status
	if status == 0 {
		status = http.StatusOK
	}

	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader(pair.Response)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Request:    req,
	}, nil
}

func (st *SnapshotTransport) save() {
	if len(st.pairs) == 0 {
		return
	}

	data, err := json.MarshalIndent(st.pairs, "", "  ")
	if err != nil {
		st.t.Fatalf("failed to marshal snapshot %s: %v", st.filePath, err)
	}

	if err := os.WriteFile(st.filePath, data, 0o644); err != nil {
		st.t.Fatalf("failed to write snapshot %s: %v", st.filePath, err)
	}
}

func (st *SnapshotTransport) RequestsRemaining() int {
	st.mu.Lock()
	defer st.mu.Unlock()
	return len(st.pairs) - st.index
}

func WriteSnapshotFile(name string, pairs []RequestResponsePair) error {
	dir := snapshotDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(pairs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, name+".json"), data, 0o644)
}

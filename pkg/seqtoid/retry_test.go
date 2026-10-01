package seqtoid

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"
)

type scriptedResponse struct {
	status     int
	retryAfter string
	body       string
}

// scriptedHTTPClient returns the scripted responses in order (repeating the last one) and records
// the request body it received on every attempt.
type scriptedHTTPClient struct {
	responses []scriptedResponse
	bodies    []string
}

func (c *scriptedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	b := []byte{}
	if req.Body != nil {
		b, _ = io.ReadAll(req.Body)
	}
	c.bodies = append(c.bodies, string(b))
	r := c.responses[len(c.responses)-1]
	if len(c.bodies) <= len(c.responses) {
		r = c.responses[len(c.bodies)-1]
	}
	header := http.Header{}
	if r.retryAfter != "" {
		header.Set("Retry-After", r.retryAfter)
	}
	return &http.Response{StatusCode: r.status, Header: header, Body: io.NopCloser(bytes.NewReader([]byte(r.body)))}, nil
}

func recordSleeps(t *testing.T) *[]time.Duration {
	waits := []time.Duration{}
	orig := sleep
	sleep = func(d time.Duration) { waits = append(waits, d) }
	t.Cleanup(func() { sleep = orig })
	return &waits
}

func TestRetriesThrottledResponseHonoringRetryAfter(t *testing.T) {
	waits := recordSleeps(t)
	hc := &scriptedHTTPClient{responses: []scriptedResponse{
		{status: 503, retryAfter: "5"},
		{status: 200, body: `{"access_key_id": "akid", "expiration": "2021-06-01T00:00:00Z"}`},
	}}
	apiClient := Client{auth0: &mockAuth0Client{}, httpClient: hc}

	creds, err := apiClient.GetUploadCredentials(1)
	if err != nil {
		t.Fatal(err)
	}
	if creds.AccessKeyID != "akid" {
		t.Errorf("access key %q, want akid", creds.AccessKeyID)
	}
	if len(hc.bodies) != 2 {
		t.Fatalf("got %d attempts, want 2", len(hc.bodies))
	}
	if hc.bodies[0] != hc.bodies[1] || hc.bodies[1] == "" {
		t.Errorf("retry did not resend the request body: %q then %q", hc.bodies[0], hc.bodies[1])
	}
	if len(*waits) != 1 || (*waits)[0] != 5*time.Second {
		t.Errorf("waits %v, want [5s]", *waits)
	}
}

func TestGivesUpAfterMaxThrottleRetries(t *testing.T) {
	waits := recordSleeps(t)
	hc := &scriptedHTTPClient{responses: []scriptedResponse{{status: 503, retryAfter: "1"}}}
	apiClient := Client{auth0: &mockAuth0Client{}, httpClient: hc}

	if _, err := apiClient.GetUploadCredentials(1); err == nil {
		t.Fatal("expected an error once retries are exhausted")
	}
	if len(hc.bodies) != maxThrottleRetries+1 {
		t.Errorf("got %d attempts, want %d", len(hc.bodies), maxThrottleRetries+1)
	}
	if len(*waits) != maxThrottleRetries {
		t.Errorf("got %d waits, want %d", len(*waits), maxThrottleRetries)
	}
}

func TestDoesNotRetryOtherErrors(t *testing.T) {
	waits := recordSleeps(t)
	hc := &scriptedHTTPClient{responses: []scriptedResponse{{status: 500}}}
	apiClient := Client{auth0: &mockAuth0Client{}, httpClient: hc}

	if _, err := apiClient.GetUploadCredentials(1); err == nil {
		t.Fatal("expected an error for HTTP 500")
	}
	if len(hc.bodies) != 1 || len(*waits) != 0 {
		t.Errorf("got %d attempts and %d waits, want 1 and 0", len(hc.bodies), len(*waits))
	}
}

func TestRetryDelay(t *testing.T) {
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	resp := func(ra string) *http.Response {
		h := http.Header{}
		if ra != "" {
			h.Set("Retry-After", ra)
		}
		return &http.Response{Header: h}
	}
	cases := []struct {
		name    string
		ra      string
		attempt int
		want    time.Duration
	}{
		{"seconds", "7", 1, 7 * time.Second},
		{"http-date", now.Add(12 * time.Second).Format(http.TimeFormat), 1, 12 * time.Second},
		{"past http-date", now.Add(-time.Minute).Format(http.TimeFormat), 1, 0},
		{"huge value capped", "3600", 1, maxRetryWait},
		{"backoff 1", "", 1, 2 * time.Second},
		{"backoff 2", "", 2, 4 * time.Second},
		{"backoff 3", "", 3, 8 * time.Second},
		{"backoff capped", "", 10, maxRetryWait},
		{"garbage falls back to backoff", "soon", 2, 4 * time.Second},
	}
	for _, tc := range cases {
		if got := retryDelay(resp(tc.ra), tc.attempt, now); got != tc.want {
			t.Errorf("%s: retryDelay = %s, want %s", tc.name, got, tc.want)
		}
	}
}

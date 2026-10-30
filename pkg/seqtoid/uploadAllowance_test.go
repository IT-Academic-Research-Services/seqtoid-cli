package seqtoid

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

// statusHTTPClient answers every request with a fixed status and body.
type statusHTTPClient struct {
	status int
	body   string
	calls  []*http.Request
}

func (c *statusHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.calls = append(c.calls, req)
	return &http.Response{StatusCode: c.status, Body: io.NopCloser(bytes.NewReader([]byte(c.body)))}, nil
}

const allowanceJSON = `{"week_start":"2026-10-25","resets_at":"2026-11-01T00:00:00-07:00","unlimited":false,
"limits":{"web":200,"cli":1000,"s3":null,"basespace":null,"total":1200},
"used":{"web":0,"cli":990,"s3":0,"basespace":0,"total":990},
"remaining":{"web":200,"cli":10,"s3":10,"basespace":10}}`

func TestGetUploadAllowance(t *testing.T) {
	httpClient := &statusHTTPClient{status: 200, body: allowanceJSON}
	apiClient := Client{auth0: &mockAuth0Client{}, httpClient: httpClient}

	allowance, err := apiClient.GetUploadAllowance()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if got := *allowance.Remaining["cli"]; got != 10 {
		t.Errorf("expected 10 CLI samples left, got %d", got)
	}
	if path := httpClient.calls[0].URL.Path; path != "/samples/upload_allowance.json" {
		t.Errorf("unexpected path %q", path)
	}
	if auth := httpClient.calls[0].Header.Get("Authorization"); !strings.HasPrefix(auth, "Bearer ") {
		t.Errorf("expected a bearer token, got %q", auth)
	}
}

// A server without the endpoint answers 401 (the path falls through to a sample route) or 404. That
// must come back as an error -- NOT exit with "not authenticated" like the other API calls -- so the
// upload carries on without a pre-check.
func TestGetUploadAllowanceOlderServer(t *testing.T) {
	for _, status := range []int{401, 403, 404, 500} {
		httpClient := &statusHTTPClient{status: status, body: `{"error":"nope"}`}
		apiClient := Client{auth0: &mockAuth0Client{}, httpClient: httpClient}
		if _, err := apiClient.GetUploadAllowance(); err == nil {
			t.Errorf("HTTP %d: expected an error", status)
		}
	}
}

func TestGetUploadAllowanceUnexpectedBody(t *testing.T) {
	for _, body := range []string{`<html>login</html>`, `{"id":1,"name":"a sample"}`} {
		httpClient := &statusHTTPClient{status: 200, body: body}
		apiClient := Client{auth0: &mockAuth0Client{}, httpClient: httpClient}
		if _, err := apiClient.GetUploadAllowance(); err == nil {
			t.Errorf("body %q: expected an error", body)
		}
	}
}

func intPtr(n int) *int { return &n }

func testAllowance(cliLeft *int) UploadAllowance {
	return UploadAllowance{
		ResetsAt:  "2026-11-01T00:00:00-07:00",
		Limits:    map[string]*int{"web": intPtr(200), "cli": intPtr(1000), "s3": nil, "basespace": nil, "total": intPtr(1200)},
		Remaining: map[string]*int{"web": intPtr(200), "cli": cliLeft, "s3": cliLeft, "basespace": cliLeft},
	}
}

func TestCheckUploadAllowanceFits(t *testing.T) {
	message, ok := CheckUploadAllowance(testAllowance(intPtr(10)), 10)
	if !ok {
		t.Fatal("expected an upload of exactly the remaining allowance to fit")
	}
	want := "Weekly upload allowance: 10 more samples this week with the SeqToID CLI, S3 import and BaseSpace import combined " +
		"(this upload uses 10). The limit resets on Sunday, November 1 at 12:00 AM Pacific."
	if message != want {
		t.Errorf("message:\n got %q\nwant %q", message, want)
	}
}

func TestCheckUploadAllowanceOver(t *testing.T) {
	message, ok := CheckUploadAllowance(testAllowance(intPtr(1)), 3)
	if ok {
		t.Fatal("expected an upload over the allowance to be refused")
	}
	want := "Weekly upload limit: this upload has 3 samples, but you can upload only 1 more sample this week " +
		"with the SeqToID CLI, S3 import and BaseSpace import combined. Nothing was uploaded. Upload fewer samples, " +
		"or the rest after the limit resets on Sunday, November 1 at 12:00 AM Pacific. Contact us if you need a higher limit."
	if message != want {
		t.Errorf("message:\n got %q\nwant %q", message, want)
	}
}

func TestCheckUploadAllowanceNoLimit(t *testing.T) {
	if message, ok := CheckUploadAllowance(testAllowance(nil), 5000); !ok || message != "" {
		t.Errorf("no CLI limit: expected silent pass, got ok=%v message=%q", ok, message)
	}
	unlimited := testAllowance(intPtr(0))
	unlimited.Unlimited = true
	if message, ok := CheckUploadAllowance(unlimited, 5000); !ok || message != "" {
		t.Errorf("unlimited user: expected silent pass, got ok=%v message=%q", ok, message)
	}
}

func TestCheckUploadAllowanceOwnS3AndBaseSpaceLimits(t *testing.T) {
	allowance := testAllowance(intPtr(500))
	allowance.Limits["s3"] = intPtr(100)
	allowance.Limits["basespace"] = intPtr(100)
	message, _ := CheckUploadAllowance(allowance, 1)
	if !strings.Contains(message, "this week with the SeqToID CLI (this upload") {
		t.Errorf("expected the CLI-only label, got %q", message)
	}
}

func TestFormatAllowanceReset(t *testing.T) {
	// After DST ends (PST, -08:00).
	if got := formatAllowanceReset("2026-11-08T00:00:00-08:00"); got != "on Sunday, November 8 at 12:00 AM Pacific" {
		t.Errorf("got %q", got)
	}
	if got := formatAllowanceReset("garbage"); got != "at the start of next week" {
		t.Errorf("got %q", got)
	}
}

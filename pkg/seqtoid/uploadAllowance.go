package seqtoid

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// UploadAllowance is GET /samples/upload_allowance.json: how many samples the user can still upload
// this week, per upload method. Remaining["cli"] is what a CLI upload may create; nil means no limit.
// It already accounts for S3/BaseSpace sharing the CLI allowance and for the weekly total.
type UploadAllowance struct {
	WeekStart string          `json:"week_start"`
	ResetsAt  string          `json:"resets_at"`
	Unlimited bool            `json:"unlimited"`
	Limits    map[string]*int `json:"limits"`
	Used      map[string]int  `json:"used"`
	Remaining map[string]*int `json:"remaining"`
}

var errUploadAllowanceUnavailable = errors.New("upload allowance unavailable")

// GetUploadAllowance fetches the user's weekly upload allowance. Unlike the other API calls it never
// exits the program: a server that predates the endpoint answers 401/404 (the path falls through to a
// sample route), and that must not look like "not authenticated" or stop an upload. Any failure is
// returned as an error and the caller carries on without a pre-check -- the server enforces the limit
// regardless.
func (c *Client) GetUploadAllowance() (UploadAllowance, error) {
	var allowance UploadAllowance

	req, err := http.NewRequest("GET", "/samples/upload_allowance.json", nil)
	if err != nil {
		return allowance, err
	}
	req.Header.Add("Accept", "application/json")
	if err := c.authorize(req); err != nil {
		return allowance, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return allowance, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return allowance, fmt.Errorf("%w: HTTP %d", errUploadAllowanceUnavailable, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return allowance, err
	}
	if err := json.Unmarshal(body, &allowance); err != nil {
		return allowance, fmt.Errorf("%w: %s", errUploadAllowanceUnavailable, err)
	}
	if allowance.ResetsAt == "" || (allowance.Remaining == nil && !allowance.Unlimited) {
		return allowance, fmt.Errorf("%w: unexpected response", errUploadAllowanceUnavailable)
	}
	return allowance, nil
}

// CheckUploadAllowance compares a CLI upload of sampleCount samples with the allowance. It returns
// the message to print (empty when the user has no CLI limit) and whether the upload fits. The
// wording matches the server's own rejection ("Weekly upload limit reached: ...").
func CheckUploadAllowance(allowance UploadAllowance, sampleCount int) (string, bool) {
	if allowance.Unlimited || allowance.Remaining == nil || allowance.Remaining["cli"] == nil {
		return "", true
	}
	left := *allowance.Remaining["cli"]
	if left < 0 {
		left = 0
	}
	label := uploadAllowanceLabel(allowance)
	resets := formatAllowanceReset(allowance.ResetsAt)

	if sampleCount > left {
		return fmt.Sprintf(
			"Weekly upload limit: this upload has %d %s, but you can upload only %d more %s this week %s. "+
				"Nothing was uploaded. Upload fewer samples, or the rest after the limit resets %s. "+
				"Contact us if you need a higher limit.",
			sampleCount, plural(sampleCount, "sample"), left, plural(left, "sample"), label, resets,
		), false
	}
	return fmt.Sprintf(
		"Weekly upload allowance: %d more %s this week %s (this upload uses %d). The limit resets %s.",
		left, plural(left, "sample"), label, sampleCount, resets,
	), true
}

// The CLI allowance is shared with S3 and BaseSpace imports unless they have limits of their own.
func uploadAllowanceLabel(allowance UploadAllowance) string {
	shared := []string{}
	for _, channel := range []string{"s3", "basespace"} {
		if allowance.Limits == nil || allowance.Limits[channel] == nil {
			shared = append(shared, map[string]string{"s3": "S3 import", "basespace": "BaseSpace import"}[channel])
		}
	}
	if len(shared) == 0 {
		return "with the SeqToID CLI"
	}
	return "with the SeqToID CLI, " + strings.Join(shared, " and ") + " combined"
}

// "on Sunday, November 1 at 12:00 AM Pacific"; the week resets at midnight Pacific.
func formatAllowanceReset(resetsAt string) string {
	t, err := time.Parse(time.RFC3339, resetsAt)
	if err != nil {
		return "at the start of next week"
	}
	if loc, err := time.LoadLocation("America/Los_Angeles"); err == nil {
		t = t.In(loc)
	}
	return "on " + t.Format("Monday, January 2 at 3:04 PM") + " Pacific"
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}

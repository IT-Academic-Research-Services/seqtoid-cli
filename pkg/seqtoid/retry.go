package seqtoid

import (
	"net/http"
	"strconv"
	"time"
)

// The SeqToID API answers 503 Service Unavailable (or 429 Too Many Requests) with a Retry-After
// header when it is briefly overloaded, e.g. while minting upload credentials for a large batch.
// Those responses mean "try again shortly", not "this request is invalid", so the CLI waits and
// retries them instead of failing -- otherwise one throttled sample aborts the rest of a batch.
const (
	maxThrottleRetries = 5
	initialRetryWait   = 2 * time.Second
	maxRetryWait       = 60 * time.Second
)

// sleep is a variable so tests can record waits instead of sleeping.
var sleep = time.Sleep

func isThrottled(statusCode int) bool {
	return statusCode == http.StatusServiceUnavailable || statusCode == http.StatusTooManyRequests
}

// retryDelay returns how long to wait before retry number `attempt` (1-based). It honors the
// server's Retry-After header (delay-seconds or an HTTP-date) and otherwise backs off
// exponentially from initialRetryWait. The result is always between 0 and maxRetryWait.
func retryDelay(res *http.Response, attempt int, now time.Time) time.Duration {
	if ra := res.Header.Get("Retry-After"); ra != "" {
		if secs, err := strconv.Atoi(ra); err == nil && secs >= 0 {
			return capRetryWait(time.Duration(secs) * time.Second)
		}
		if at, err := http.ParseTime(ra); err == nil {
			return capRetryWait(at.Sub(now))
		}
	}
	wait := initialRetryWait
	for i := 1; i < attempt && wait < maxRetryWait; i++ {
		wait *= 2
	}
	return capRetryWait(wait)
}

func capRetryWait(d time.Duration) time.Duration {
	if d < 0 {
		return 0
	}
	if d > maxRetryWait {
		return maxRetryWait
	}
	return d
}

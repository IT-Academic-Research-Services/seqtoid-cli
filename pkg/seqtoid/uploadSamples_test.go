package seqtoid

import (
	"encoding/json"
	"io"
	"strings"
	"testing"
)

// captureCreateSamplesBody runs CreateSamples against a mock HTTP client and returns the
// JSON request body the CLI sent to /samples/bulk_upload_with_metadata.
func captureCreateSamplesBody(t *testing.T, workflow string, opts SampleOptions) samplesReq {
	t.Helper()

	httpClient := newMockHTTPClient([]byte(`{"samples":[],"errors":[]}`))
	apiClient := Client{
		auth0:      &mockAuth0Client{},
		httpClient: &httpClient,
	}

	metadata := SamplesMetadata{
		"sample1": Metadata{HostGenome: "Human"},
	}
	sampleFiles := map[string]SampleFiles{
		"sample1": {Single: []string{"sample1.fastq.gz"}},
	}

	if _, err := apiClient.CreateSamples(1, sampleFiles, metadata, workflow, opts); err != nil {
		t.Fatalf("CreateSamples returned error: %v", err)
	}

	if len(httpClient.calls) != 1 {
		t.Fatalf("expected exactly 1 request, got %d", len(httpClient.calls))
	}

	bodyBytes, err := io.ReadAll(httpClient.calls[0].Body)
	if err != nil {
		t.Fatalf("could not read request body: %v", err)
	}

	var req samplesReq
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		t.Fatalf("could not unmarshal request body %q: %v", string(bodyBytes), err)
	}
	return req
}

// The long-read (Nanopore/ONT) metagenomics path must forward guppy_basecaller_setting to the
// server; it is a required input to the long-read mNGS WDL. Regression guard for the bug where
// the CLI collected --guppy-basecaller-setting but never serialized it.
func TestCreateSamplesSerializesGuppyForLongRead(t *testing.T) {
	req := captureCreateSamplesBody(t, "long-read-mngs", SampleOptions{
		Technology:             "ONT",
		GuppyBasecallerSetting: "hac",
	})

	if len(req.Samples) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(req.Samples))
	}
	got := req.Samples[0].GuppyBasecallerSetting
	if got == nil {
		t.Fatal("guppy_basecaller_setting was nil; expected it to be forwarded for long-read-mngs")
	}
	if *got != "hac" {
		t.Errorf("guppy_basecaller_setting = %q, want %q", *got, "hac")
	}
	if req.Samples[0].Technology != "ONT" {
		t.Errorf("technology = %q, want %q", req.Samples[0].Technology, "ONT")
	}
}

// The Illumina (short-read) path must NOT send guppy_basecaller_setting -- the server rejects it
// for Illumina. omitempty on a nil pointer keeps the key out of the JSON entirely.
func TestCreateSamplesOmitsGuppyForShortRead(t *testing.T) {
	req := captureCreateSamplesBody(t, "short-read-mngs", SampleOptions{
		Technology: "Illumina",
	})

	if len(req.Samples) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(req.Samples))
	}
	if req.Samples[0].GuppyBasecallerSetting != nil {
		t.Errorf("guppy_basecaller_setting = %v, want nil for Illumina", *req.Samples[0].GuppyBasecallerSetting)
	}

	// Belt-and-suspenders: the raw JSON must not carry the key at all.
	raw, err := json.Marshal(req.Samples[0])
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if strings.Contains(string(raw), "guppy_basecaller_setting") {
		t.Errorf("serialized Illumina sample unexpectedly contains guppy_basecaller_setting: %s", string(raw))
	}
}

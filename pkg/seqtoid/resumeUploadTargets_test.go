package seqtoid

import (
	"encoding/json"
	"io"
	"testing"
)

func TestResumeUploadTargets(t *testing.T) {
	response := []byte(`{
		"samples": [{"id": 7, "name": "S1", "input_files": [
			{"multipart_upload_id": "mpu-1", "s3_path": "s3://b/samples/1/7/fastqs/x_R1.fastq.gz", "source": "S1_R1.fastq.gz", "uploaded": true},
			{"multipart_upload_id": null, "s3_path": "s3://b/samples/1/7/fastqs/x_R2.fastq.gz", "source": "S1_R2.fastq.gz", "uploaded": false}
		]}],
		"errors": {"S2": "already uploaded"}
	}`)
	httpClient := newMockHTTPClient(response)
	apiClient := Client{auth0: &mockAuth0Client{}, httpClient: &httpClient}

	samples, skipped, err := apiClient.ResumeUploadTargets(123, []string{"S1", "S2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(samples) != 1 || samples[0].ID != 7 || samples[0].Name != "S1" {
		t.Fatalf("unexpected samples: %+v", samples)
	}
	if !samples[0].InputFiles[0].Uploaded || samples[0].InputFiles[1].Uploaded {
		t.Errorf("uploaded flags not parsed: %+v", samples[0].InputFiles)
	}
	if skipped["S2"] != "already uploaded" {
		t.Errorf("errors not parsed: %v", skipped)
	}

	if len(httpClient.calls) != 1 {
		t.Fatalf("expected 1 request, got %d", len(httpClient.calls))
	}
	req := httpClient.calls[0]
	if req.Method != "POST" || req.URL.Path != "/samples/resume_upload_targets.json" {
		t.Errorf("unexpected request %s %s", req.Method, req.URL.Path)
	}
	raw, _ := io.ReadAll(req.Body)
	var body resumeUploadTargetsReq
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("request body is not JSON: %v (%s)", err, raw)
	}
	if body.ProjectID != 123 || len(body.SampleNames) != 2 {
		t.Errorf("unexpected request body: %+v", body)
	}
}

func TestPendingInputFilesSkipsUploaded(t *testing.T) {
	sample := createSamplesResSample{InputFiles: []UploadInfo{
		{S3Path: "a", Uploaded: true},
		{S3Path: "b", Uploaded: false},
		{S3Path: "c"},
	}}
	pending := pendingInputFiles(sample)
	if len(pending) != 2 || pending[0].S3Path != "b" || pending[1].S3Path != "c" {
		t.Errorf("unexpected pending files: %+v", pending)
	}
}

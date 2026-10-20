package seqtoid

type resumeUploadTargetsReq struct {
	ProjectID   int      `json:"project_id"`
	SampleNames []string `json:"sample_names"`
}

type resumeUploadTargetsRes struct {
	Samples []createSamplesResSample `json:"samples"`
	Errors  map[string]string        `json:"errors"`
}

// ResumeUploadTargets re-attaches to samples a previous upload created but never finished (the client
// died, slept or lost its network mid-batch). For each named sample the user uploaded to the project,
// the server returns the same upload targets CreateSamples returns, plus whether each file is already
// on S3. Names that cannot be resumed come back in the errors map with the server's reason.
func (c *Client) ResumeUploadTargets(projectID int, sampleNames []string) ([]createSamplesResSample, map[string]string, error) {
	var res resumeUploadTargetsRes
	err := c.request(
		"POST",
		"/samples/resume_upload_targets.json",
		"",
		resumeUploadTargetsReq{ProjectID: projectID, SampleNames: sampleNames},
		&res,
	)
	return res.Samples, res.Errors, err
}

// pendingInputFiles is the input files of a sample that still need uploading: everything except the
// files the server reports as already on S3 (only a resumed sample carries that flag).
func pendingInputFiles(sample createSamplesResSample) []UploadInfo {
	pending := make([]UploadInfo, 0, len(sample.InputFiles))
	for _, f := range sample.InputFiles {
		if !f.Uploaded {
			pending = append(pending, f)
		}
	}
	return pending
}

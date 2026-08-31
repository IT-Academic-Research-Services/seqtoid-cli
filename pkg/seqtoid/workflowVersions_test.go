package seqtoid

import "testing"

func TestGetWorkflowVersions(t *testing.T) {
	response := []byte(`{"workflow":"short-read-mngs","versions":[{"version":"8.3.1","deprecated":false},{"version":"8.2.0","deprecated":true,"notes":"old"}]}`)
	httpClient := newMockHTTPClient(response)
	apiClient := Client{
		auth0:      &mockAuth0Client{},
		httpClient: &httpClient,
	}

	versions, err := apiClient.GetWorkflowVersions("short-read-mngs")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
	if versions[0].Version != "8.3.1" || versions[0].Deprecated {
		t.Errorf("unexpected first version: %+v", versions[0])
	}
	if versions[1].Version != "8.2.0" || !versions[1].Deprecated || versions[1].Notes != "old" {
		t.Errorf("unexpected second version: %+v", versions[1])
	}

	if len(httpClient.calls) != 1 {
		t.Fatalf("expected 1 request, got %d", len(httpClient.calls))
	}
	if got := httpClient.calls[0].URL.Query().Get("workflow"); got != "short-read-mngs" {
		t.Errorf("expected workflow query param, got %q", got)
	}
}

func TestValidateWorkflowVersionAccepted(t *testing.T) {
	response := []byte(`{"workflow":"amr","versions":[{"version":"1.4.0","deprecated":false}]}`)
	httpClient := newMockHTTPClient(response)
	apiClient := Client{
		auth0:      &mockAuth0Client{},
		httpClient: &httpClient,
	}

	if err := apiClient.ValidateWorkflowVersion("amr", "1.4.0"); err != nil {
		t.Fatalf("expected version to validate, got error: %s", err)
	}
}

func TestValidateWorkflowVersionRejectsUnavailable(t *testing.T) {
	response := []byte(`{"workflow":"amr","versions":[{"version":"1.4.0","deprecated":false}]}`)
	httpClient := newMockHTTPClient(response)
	apiClient := Client{
		auth0:      &mockAuth0Client{},
		httpClient: &httpClient,
	}

	if err := apiClient.ValidateWorkflowVersion("amr", "9.9.9"); err == nil {
		t.Fatal("expected an error for an unavailable version")
	}
}

func TestValidateWorkflowVersionRejectsMalformed(t *testing.T) {
	// A malformed value is rejected before any request is made.
	httpClient := newMockHTTPClient([]byte(`{}`))
	apiClient := Client{
		auth0:      &mockAuth0Client{},
		httpClient: &httpClient,
	}

	if err := apiClient.ValidateWorkflowVersion("amr", "not-a-version"); err == nil {
		t.Fatal("expected an error for a malformed version")
	}
	if len(httpClient.calls) != 0 {
		t.Errorf("expected no request for a malformed version, got %d", len(httpClient.calls))
	}
}

func TestValidateWorkflowVersionSkipsWhenCatalogEmpty(t *testing.T) {
	// An empty catalog means per-run selection is off; pass through without error.
	response := []byte(`{"workflow":"amr","versions":[]}`)
	httpClient := newMockHTTPClient(response)
	apiClient := Client{
		auth0:      &mockAuth0Client{},
		httpClient: &httpClient,
	}

	if err := apiClient.ValidateWorkflowVersion("amr", "1.4.0"); err != nil {
		t.Fatalf("expected no error when catalog is empty, got: %s", err)
	}
}

package seqtoid

import (
	"fmt"
	"net/url"
	"regexp"
)

// CatalogedWorkflowVersion is one selectable pipeline version for a workflow, as returned by the
// server's workflow-version catalog. Deprecated versions are still runnable but no longer patched.
type CatalogedWorkflowVersion struct {
	Version    string `json:"version"`
	Deprecated bool   `json:"deprecated"`
	Notes      string `json:"notes"`
}

type workflowVersionsResp struct {
	Workflow string                     `json:"workflow"`
	Versions []CatalogedWorkflowVersion `json:"versions"`
}

// userVersionPrefixFormat mirrors the server's accepted shape for a user-selected version: a major
// (8), major.minor (8.1) or full (8.1.2) version. A value that does not match is rejected client
// side so the user gets a clear message instead of a server validation error.
var userVersionPrefixFormat = regexp.MustCompile(`^\d+(\.\d+){0,2}$`)

// GetWorkflowVersions fetches the catalog of selectable versions for a workflow, newest first.
//
// The server returns an empty list (not an error) when per-run version selection is disabled for
// the environment. Callers treat an empty list as "no selectable versions published" rather than a
// failure: the server uses the configured default version in that case.
func (c *Client) GetWorkflowVersions(workflow string) ([]CatalogedWorkflowVersion, error) {
	query := url.Values{"workflow": []string{workflow}}
	var resp workflowVersionsResp
	err := c.request("GET", "/workflow_versions", query.Encode(), struct{}{}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Versions, nil
}

// ValidateWorkflowVersion checks a user-supplied version against the catalog for a workflow before
// upload. Selection is literal on the server (an explicit choice runs exactly as chosen and is not
// expanded to the latest sharing its prefix), so the version must exactly match a catalogued entry.
//
// If the catalog is empty (per-run selection disabled for the environment, or nothing published),
// validation is skipped and the version is passed through unchanged: the server ignores the
// selection when the feature is off, and the default behavior is preserved.
func (c *Client) ValidateWorkflowVersion(workflow, version string) error {
	if !userVersionPrefixFormat.MatchString(version) {
		return fmt.Errorf(
			"workflow-version %q must be a major (8), major.minor (8.1) or full version (8.1.2)",
			version,
		)
	}

	versions, err := c.GetWorkflowVersions(workflow)
	if err != nil {
		return err
	}
	if len(versions) == 0 {
		// Per-run selection not available for this environment; leave the server to resolve the
		// default. Nothing to validate against.
		return nil
	}

	available := make([]string, 0, len(versions))
	for _, v := range versions {
		if v.Version == version {
			return nil
		}
		available = append(available, v.Version)
	}

	return fmt.Errorf(
		"workflow-version %q is not an available version for %s; choose one of: %v",
		version, workflow, available,
	)
}

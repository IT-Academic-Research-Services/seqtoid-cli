package seqtoid

import (
	"testing"
)

func TestGetMetadataForHost(t *testing.T) {
	response := []byte(`[{
      "display_name": "test",
      "description": "desc",
      "examples": "{\"all\": [\"example\"] }"
    }]`)
	httpClient := newMockHTTPClient(response)
	apiClient := Client{
		auth0:      &mockAuth0Client{},
		httpClient: &httpClient,
	}

	fields, err := apiClient.GetMetadataForHostGenome("human")
	if err != nil {
		t.Fatal(err)
	}

	if len(fields) != 1 {
		t.Errorf("expected 1 field but got %d", len(fields))
	}

	if fields[0].Name != "test" {
		t.Errorf("expected name to be 'test' but it was '%s'", fields[0].Name)
	}

	if len(fields[0].Example.All) != 1 {
		t.Errorf("expected one option in All but got %d", 1)
	}

	if fields[0].Example.All[0] != "example" {
		t.Errorf("expected option to be 'example' but it was '%s'", fields[0].Example.All[0])
	}
}

func TestGetMetadataForHostEmptyExamples(t *testing.T) {
	// The API returns "examples": null for fields without examples (e.g. age,
	// collection_location_v2). That decodes to an empty string; parsing it as JSON
	// used to fail with "unexpected end of JSON input" and abort the whole command.
	response := []byte(`[
    {"display_name": "with", "description": "d", "examples": "{\"all\": [\"ex\"] }"},
    {"display_name": "age",  "description": "d", "examples": null}
  ]`)
	httpClient := newMockHTTPClient(response)
	apiClient := Client{
		auth0:      &mockAuth0Client{},
		httpClient: &httpClient,
	}

	fields, err := apiClient.GetMetadataForHostGenome("human")
	if err != nil {
		t.Fatalf("empty examples must not error, got: %v", err)
	}
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields but got %d", len(fields))
	}
	if len(fields[1].Example.All) != 0 {
		t.Errorf("expected the null-examples field to have no examples, got %d", len(fields[1].Example.All))
	}
}

func TestGetMetadataForHostParsingError(t *testing.T) {
	response := []byte(`[{
      "display_name": "test",
      "description": "desc",
      "examples": "{\"all\": [\"example] }"
    }]`)
	httpClient := newMockHTTPClient(response)
	apiClient := Client{
		auth0:      &mockAuth0Client{},
		httpClient: &httpClient,
	}

	_, err := apiClient.GetMetadataForHostGenome("human")
	if err == nil {
		t.Errorf("expected error from invalid JSON  but error was nil")
	}
}

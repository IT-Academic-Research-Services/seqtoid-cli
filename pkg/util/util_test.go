package util

import "testing"

func TestParseMetadataPairs(t *testing.T) {
	m, err := ParseMetadataPairs([]string{
		"Host Organism=Human",
		"Collection Location=Santa Barbara, CA, USA", // value with commas must survive
		"Notes=a=b=c",                                // only the first '=' splits
		" Sample Type =CSF",                          // name is trimmed
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]string{
		"Host Organism":       "Human",
		"Collection Location": "Santa Barbara, CA, USA",
		"Notes":               "a=b=c",
		"Sample Type":         "CSF",
	}
	if len(m) != len(want) {
		t.Fatalf("expected %d pairs, got %d: %v", len(want), len(m), m)
	}
	for k, v := range want {
		if m[k] != v {
			t.Errorf("pair %q: expected %q, got %q", k, v, m[k])
		}
	}
}

func TestParseMetadataPairsErrors(t *testing.T) {
	for _, bad := range []string{"no-equals-sign", "=value-without-name"} {
		if _, err := ParseMetadataPairs([]string{bad}); err == nil {
			t.Errorf("expected error for %q, got nil", bad)
		}
	}
	// empty input is valid (no metadata provided)
	if _, err := ParseMetadataPairs(nil); err != nil {
		t.Errorf("nil input should not error, got %v", err)
	}
}

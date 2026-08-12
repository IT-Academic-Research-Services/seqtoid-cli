package seqtoid

import (
	"encoding/json"
	"net/url"
)

type getMetadataForHostGenomeReq struct{}

type getMetadataForHostGenomeMetadataField struct {
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Examples    string `json:"examples"`
}

type Example struct {
	All []string `json:"all"`
	One []string `json:"1"`
}

type MetadataField struct {
	Name        string
	Description string
	Example     Example
}

func (c *Client) GetMetadataForHostGenome(hostGenome string) ([]MetadataField, error) {
	query := url.Values{
		"name": []string{hostGenome},
	}

	var res []getMetadataForHostGenomeMetadataField
	err := c.request(
		"GET",
		"/metadata/metadata_for_host_genome.json",
		query.Encode(),
		getMetadataForHostGenomeReq{},
		&res,
	)
	if err != nil {
		return nil, err
	}

	metadataFields := make([]MetadataField, len(res))
	for i, f := range res {
		var ex Example
		// Not every field has examples; the server returns null for those, which
		// decodes to an empty string. Unmarshalling "" would fail ("unexpected end
		// of JSON input"), so skip parsing and leave the zero-value Example.
		if f.Examples != "" {
			if err := json.Unmarshal([]byte(f.Examples), &ex); err != nil {
				return nil, err
			}
		}
		metadataFields[i] = MetadataField{
			Name:        f.DisplayName,
			Description: f.Description,
			Example:     ex,
		}
	}

	return metadataFields, nil
}

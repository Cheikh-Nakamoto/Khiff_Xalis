// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package source

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// PoliticalRiskSource queries the GDELT API for political violence events.
type PoliticalRiskSource struct {
	http *HTTPClient
}

// NewPoliticalRiskSource creates a PoliticalRiskSource.
func NewPoliticalRiskSource(http *HTTPClient) *PoliticalRiskSource {
	return &PoliticalRiskSource{http: http}
}

// FetchGDELTEvents counts political violence articles/events in the past 7 days
// for the given GDELT country name.
func (s *PoliticalRiskSource) FetchGDELTEvents(ctx context.Context, gdeltCountryName string) (int, error) {
	url := fmt.Sprintf(GDELTPoliticalViolenceURL, gdeltCountryName)

	resp, err := s.http.Get(ctx, url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("GDELT status code %d", resp.StatusCode)
	}

	var result struct {
		EventCount int `json:"eventcount"`
		// GDELT summary payload might contain an events array.
		Events []interface{} `json:"events"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// GDELT sometimes returns plain text on no events; handle gracefully.
		return 0, nil
	}

	// Fallback to length of events slice if eventcount is not populated.
	if result.EventCount == 0 && len(result.Events) > 0 {
		return len(result.Events), nil
	}

	return result.EventCount, nil
}

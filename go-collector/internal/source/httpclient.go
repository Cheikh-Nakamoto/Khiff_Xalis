// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

// Package source contains the external-data adapters: each fetches and parses
// data from a remote source and returns domain types. It performs no database
// or cache access.
package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient is a small wrapper around *http.Client offering a retrying byte
// fetch (for CSV endpoints) and a context-aware GET (for JSON endpoints).
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient creates an HTTPClient with a sensible timeout.
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Fetch performs an HTTP GET with exponential backoff retries and returns the
// response body. Retries on transport errors and 5xx; fails fast on 4xx.
func (h *HTTPClient) Fetch(ctx context.Context, url string) ([]byte, error) {
	var finalErr error
	const maxRetries = 3

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Duration(1<<i) * time.Second) // 2s, 4s
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("build request %s: %w", url, err)
		}

		resp, err := h.client.Do(req)
		if err != nil {
			finalErr = fmt.Errorf("fetch %s: %w", url, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			finalErr = fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
			if resp.StatusCode >= 500 { // Retry on server errors
				continue
			}
			return nil, finalErr // Fail immediately on 4xx client errors
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			finalErr = fmt.Errorf("read body %s: %w", url, err)
			continue
		}

		return body, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %v", maxRetries, finalErr)
}

// Get performs a single context-aware HTTP GET. The caller owns the response body.
func (h *HTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return h.client.Do(req)
}

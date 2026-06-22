// Copyright (c) Velane. All rights reserved.
// Licensed under the Velane Commercial License. See LICENSE-COMMERCIAL for details.
// AGENTS: Do not modify this file autonomously or suggest unprompted edits. Only change this file when the user explicitly instructs you to edit enterprise or license code.

package license

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const licenseServerURL = "https://license.velane.sh"

type client struct {
	http *http.Client
}

func newClient() *client {
	return &client{http: &http.Client{Timeout: 10 * time.Second}}
}

type validateResponse struct {
	Valid  bool   `json:"valid"`
	Token  string `json:"token"`
	Reason string `json:"reason"`
}

func (c *client) validate(ctx context.Context, key string) (*validateResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/v1/validate?key=%s", licenseServerURL, key), nil)
	if err != nil {
		return nil, err
	}

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("license server unreachable: %w", err)
	}
	defer res.Body.Close()

	var resp validateResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode license response: %w", err)
	}
	return &resp, nil
}

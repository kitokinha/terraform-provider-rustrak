package alertchannel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/kitokinha/terraform-provider-rustrak/internal/client"
)

type AlertChannel struct {
	ID           int             `json:"id"`
	Name         string          `json:"name"`
	ProviderType string          `json:"provider_type"`
	Credentials  json.RawMessage `json:"credentials"`
	IsEnabled    bool            `json:"is_enabled"`
}

type CreateAlertChannelRequest struct {
	Name         string          `json:"name"`
	ProviderType string          `json:"provider_type"`
	Credentials  json.RawMessage `json:"credentials"`
	IsEnabled    bool            `json:"is_enabled"`
}

type UpdateAlertChannelRequest struct {
	Name        string          `json:"name"`
	Credentials json.RawMessage `json:"credentials,omitempty"`
	IsEnabled   bool            `json:"is_enabled"`
}

func request(c *client.RustrakClient, ctx context.Context, method, path string, body, result any) error {
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.HostURL, "/")+path, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	data, err = c.Do(req)
	if err != nil {
		// API error bodies can echo credentials; keep them out of Terraform diagnostics.
		if apiErr, ok := err.(*client.APIError); ok {
			return &client.APIError{StatusCode: apiErr.StatusCode, Status: apiErr.Status}
		}
		return err
	}
	if result != nil {
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func ReadAlertChannel(c *client.RustrakClient, ctx context.Context, id string) (*AlertChannel, error) {
	var channel AlertChannel
	err := request(c, ctx, http.MethodGet, "/api/integrations/"+url.PathEscape(id), nil, &channel)
	if client.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func CreateAlertChannel(c *client.RustrakClient, ctx context.Context, input CreateAlertChannelRequest) (*AlertChannel, error) {
	var channel AlertChannel
	if err := request(c, ctx, http.MethodPost, "/api/integrations", input, &channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

func UpdateAlertChannel(c *client.RustrakClient, ctx context.Context, id string, input UpdateAlertChannelRequest) (*AlertChannel, error) {
	var channel AlertChannel
	if err := request(c, ctx, http.MethodPatch, "/api/integrations/"+url.PathEscape(id), input, &channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

func DeleteAlertChannel(c *client.RustrakClient, ctx context.Context, id string) error {
	err := request(c, ctx, http.MethodDelete, "/api/integrations/"+url.PathEscape(id), nil, nil)
	if client.IsNotFound(err) {
		return nil
	}
	return err
}

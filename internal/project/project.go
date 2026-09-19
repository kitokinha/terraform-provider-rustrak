package project

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kitokinha/terraform-provider-rustrak/internal/client"
)

type Project struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	DSN      string `json:"dsn"`
	Platform string `json:"platform,omitempty"`
}

type CreateProjectRequest struct {
	Name     string `json:"name"`
	Platform string `json:"platform,omitempty"`
	Slug     string `json:"slug,omitempty"`
}

type UpdateProjectRequest struct {
	Name     string `json:"name,omitempty"`
	Platform string `json:"platform,omitempty"`
	Slug     string `json:"slug,omitempty"`
}

func request(
	c *client.RustrakClient,
	ctx context.Context,
	method string,
	path string,
	body any,
	result any,
) error {
	var reqBody *strings.Reader

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}

		reqBody = strings.NewReader(string(data))
	} else {
		reqBody = strings.NewReader("")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		c.HostURL+path,
		reqBody,
	)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("rustrak API returned %s", resp.Status)
	}

	if result == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func ReadProject(
	c *client.RustrakClient,
	ctx context.Context,
	id string,
) (*Project, error) {
	var project Project

	err := request(
		c,
		ctx,
		http.MethodGet,
		"/api/projects/"+id,
		nil,
		&project,
	)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func CreateProject(
	c *client.RustrakClient,
	ctx context.Context,
	input CreateProjectRequest,
) (*Project, error) {
	var project Project

	err := request(
		c,
		ctx,
		http.MethodPost,
		"/api/projects",
		input,
		&project,
	)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func UpdateProject(
	c *client.RustrakClient,
	ctx context.Context,
	id string,
	input UpdateProjectRequest,
) (*Project, error) {
	var project Project

	err := request(
		c,
		ctx,
		http.MethodPut,
		"/api/projects/"+id,
		input,
		&project,
	)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func DeleteProject(
	c *client.RustrakClient,
	ctx context.Context,
	id string,
) error {
	return request(
		c,
		ctx,
		http.MethodDelete,
		"/api/projects/"+id,
		nil,
		nil,
	)
}

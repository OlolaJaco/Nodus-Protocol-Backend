package payments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client talks to the Nodus Core Engine over HTTP.
type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) InitiatePayment(ctx context.Context, req initiateRequest) (*enginePayment, error) {
	var out enginePayment
	return &out, c.post(ctx, "/api/v1/payments", req, &out)
}

func (c *Client) GetPayment(ctx context.Context, engineID string) (*enginePayment, error) {
	var out enginePayment
	return &out, c.get(ctx, "/api/v1/payments/"+engineID, &out)
}

func (c *Client) ListPayments(ctx context.Context) ([]enginePayment, error) {
	var out []enginePayment
	return out, c.get(ctx, "/api/v1/payments", &out)
}

func (c *Client) SimulatePayment(ctx context.Context, req simulateRequest) (*engineSimulation, error) {
	var out engineSimulation
	return &out, c.post(ctx, "/api/v1/payments/simulate", req, &out)
}

func (c *Client) BatchPayments(ctx context.Context, items []batchItem) (*engineBatchResult, error) {
	var out engineBatchResult
	return &out, c.post(ctx, "/api/v1/payments/batch", items, &out)
}

func (c *Client) GetReceipt(ctx context.Context, engineID string) (*engineReceipt, error) {
	var out engineReceipt
	return &out, c.get(ctx, "/api/v1/payments/"+engineID+"/receipt", &out)
}

func (c *Client) GetFees(ctx context.Context) ([]engineFees, error) {
	var out []engineFees
	return out, c.get(ctx, "/api/v1/fees/current", &out)
}

func (c *Client) GetRates(ctx context.Context, tokens string) ([]engineRate, error) {
	var out []engineRate
	path := "/api/v1/rates"
	if tokens != "" {
		path += "?tokens=" + tokens
	}
	return out, c.get(ctx, path, &out)
}

func (c *Client) Health(ctx context.Context) (*engineHealth, error) {
	var out engineHealth
	return &out, c.get(ctx, "/healthz", &out)
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("core engine request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("core engine returned %d: %s", resp.StatusCode, string(raw))
	}

	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

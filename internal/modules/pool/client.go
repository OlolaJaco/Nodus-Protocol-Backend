package pool

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

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) GetReserves(ctx context.Context) (*reserves, error) {
	var out reserves
	return &out, c.get(ctx, "/api/v1/pool/reserves", &out)
}

func (c *Client) GetQuote(ctx context.Context, amountIn, tokenIn string) (*priceQuote, error) {
	var out priceQuote
	path := fmt.Sprintf("/api/v1/pool/quote?amount_in=%s&token_in=%s", amountIn, tokenIn)
	return &out, c.get(ctx, path, &out)
}

func (c *Client) GetLPBalance(ctx context.Context, address string) (*lpBalance, error) {
	var out lpBalance
	return &out, c.get(ctx, "/api/v1/pool/lp-balance?address="+address, &out)
}

func (c *Client) GetStats(ctx context.Context) (*poolStats, error) {
	var out poolStats
	return &out, c.get(ctx, "/api/v1/pool/stats", &out)
}

func (c *Client) BuildSwapParams(ctx context.Context, req swapParamsRequest) (*unsignedTx, error) {
	var out unsignedTx
	return &out, c.post(ctx, "/api/v1/pool/build/swap", req, &out)
}

func (c *Client) BuildAddLiquidity(ctx context.Context, req addLiquidityRequest) (*unsignedTx, error) {
	var out unsignedTx
	return &out, c.post(ctx, "/api/v1/pool/build/add-liquidity", req, &out)
}

func (c *Client) BuildRemoveLiquidity(ctx context.Context, req removeLiquidityRequest) (*unsignedTx, error) {
	var out unsignedTx
	return &out, c.post(ctx, "/api/v1/pool/build/remove-liquidity", req, &out)
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
		return fmt.Errorf("marshal: %w", err)
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
		return fmt.Errorf("pool client: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("core engine pool returned %d: %s", resp.StatusCode, string(raw))
	}
	if out != nil {
		return json.Unmarshal(raw, out)
	}
	return nil
}

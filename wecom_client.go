package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type wecomClient struct {
	cfg  config
	http *http.Client
}

type accessTokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type apiErrorResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (c *wecomClient) requireCredentials() error {
	if err := c.loadGatewayToken(); err != nil {
		return err
	}
	if missing := requiredMissing(map[string]string{
		"--gateway-base-url or WECOM_GATEWAY_BASE_URL": c.cfg.GatewayBaseURL,
		"--identity-file or CLI_IDENTITY_FILE":         c.cfg.GatewayToken,
	}); len(missing) > 0 {
		return fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (c *wecomClient) requireGatewayToken() error {
	if err := c.loadGatewayToken(); err != nil {
		return err
	}
	if missing := requiredMissing(map[string]string{
		"--identity-file or CLI_IDENTITY_FILE": c.cfg.GatewayToken,
	}); len(missing) > 0 {
		return fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (c *wecomClient) loadGatewayToken() error {
	if strings.TrimSpace(c.cfg.GatewayToken) == "" && strings.TrimSpace(c.cfg.IdentityFile) != "" {
		token, err := readAccessToken(c.cfg.IdentityFile)
		if err != nil {
			return err
		}
		c.cfg.GatewayToken = token
	}
	return nil
}

func (c *wecomClient) postWeCom(path string, body any) error {
	raw, err := c.postWeComRaw(path, body)
	if err != nil {
		return err
	}
	return printRawResponse(raw)
}

func (c *wecomClient) postWeComRaw(path string, body any) ([]byte, error) {
	rawBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}
	req, err := c.newGatewayPostRequest(path, rawBody)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("POST %s returned HTTP %d: %s", path, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var apiErr apiErrorResponse
	if json.Unmarshal(raw, &apiErr) == nil && apiErr.ErrCode != 0 {
		return nil, fmt.Errorf("WeCom returned errcode %d: %s", apiErr.ErrCode, apiErr.ErrMsg)
	}
	return raw, nil
}

func (c *wecomClient) newGatewayPostRequest(path string, rawBody []byte) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, c.cfg.GatewayBaseURL+path, bytes.NewReader(rawBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.GatewayToken)
	return req, nil
}

func printRawResponse(raw []byte) error {
	if len(bytes.TrimSpace(raw)) == 0 {
		fmt.Println("{}")
		return nil
	}
	var formatted bytes.Buffer
	if json.Valid(raw) && json.Indent(&formatted, raw, "", "  ") == nil {
		fmt.Println(formatted.String())
		return nil
	}
	fmt.Println(string(raw))
	return nil
}

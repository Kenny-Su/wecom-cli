package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const tokenRefreshSkew = 5 * time.Minute

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

type cachedToken struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

type tokenCacheFile struct {
	Tokens map[string]cachedToken `json:"tokens"`
}

func (c *wecomClient) requireCredentials() error {
	if missing := requiredMissing(map[string]string{
		"--corpid or WECOM_CORP_ID":         c.cfg.CorpID,
		"--corpsecret or WECOM_CORP_SECRET": c.cfg.CorpSecret,
	}); len(missing) > 0 {
		return fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (c *wecomClient) accessToken() (string, error) {
	key := tokenCacheKey(c.cfg.CorpID, c.cfg.CorpSecret)
	if token, ok := readCachedToken(c.cfg.TokenCache, key); ok {
		return token, nil
	}
	q := url.Values{}
	q.Set("corpid", c.cfg.CorpID)
	q.Set("corpsecret", c.cfg.CorpSecret)
	req, err := http.NewRequest(http.MethodGet, defaultBaseURL+"/cgi-bin/gettoken?"+q.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("build gettoken request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("get access_token: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read gettoken response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("gettoken returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var parsed accessTokenResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("parse gettoken response: %w", err)
	}
	if parsed.ErrCode != 0 {
		return "", fmt.Errorf("gettoken returned errcode %d: %s", parsed.ErrCode, parsed.ErrMsg)
	}
	if parsed.AccessToken == "" {
		return "", errors.New("gettoken response did not include access_token")
	}
	writeCachedToken(c.cfg.TokenCache, key, parsed.AccessToken, time.Now().Add(time.Duration(parsed.ExpiresIn)*time.Second))
	return parsed.AccessToken, nil
}

func (c *wecomClient) postWeCom(path string, body any) error {
	rawBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, defaultBaseURL+path, bytes.NewReader(rawBody))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("POST %s returned HTTP %d: %s", path, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var apiErr apiErrorResponse
	if json.Unmarshal(raw, &apiErr) == nil && apiErr.ErrCode != 0 {
		return fmt.Errorf("WeCom returned errcode %d: %s", apiErr.ErrCode, apiErr.ErrMsg)
	}
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

func readCachedToken(path, key string) (string, bool) {
	raw, err := os.ReadFile(expandPath(path))
	if err != nil {
		return "", false
	}
	var cache tokenCacheFile
	if err := json.Unmarshal(raw, &cache); err != nil {
		return "", false
	}
	token, ok := cache.Tokens[key]
	if !ok || token.AccessToken == "" {
		return "", false
	}
	if time.Unix(token.ExpiresAt, 0).Before(time.Now().Add(tokenRefreshSkew)) {
		return "", false
	}
	return token.AccessToken, true
}

func writeCachedToken(path, key, token string, expiresAt time.Time) {
	path = expandPath(path)
	cache := tokenCacheFile{Tokens: map[string]cachedToken{}}
	if raw, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(raw, &cache)
	}
	if cache.Tokens == nil {
		cache.Tokens = map[string]cachedToken{}
	}
	cache.Tokens[key] = cachedToken{AccessToken: token, ExpiresAt: expiresAt.Unix()}
	raw, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return
	}
	_ = os.WriteFile(path, raw, 0o600)
}

func tokenCacheKey(corpID, corpSecret string) string {
	sum := sha256.Sum256([]byte(corpID + "\x00" + corpSecret))
	return hex.EncodeToString(sum[:])
}

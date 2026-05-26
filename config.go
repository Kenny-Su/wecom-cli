package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL     = "https://qyapi.weixin.qq.com"
	defaultHTTPTimeout = 30 * time.Second
)

type config struct {
	CorpID        string
	CorpSecret    string
	BaseURL       string
	TokenCache    string
	ResourceTable string
	HTTPClient    *http.Client
}

func parseGlobalFlags(args []string) (config, []string, error) {
	if err := loadDotEnv(".env"); err != nil {
		return config{}, nil, err
	}

	cfg := config{
		CorpID:        strings.TrimSpace(os.Getenv("WECOM_CORP_ID")),
		CorpSecret:    strings.TrimSpace(os.Getenv("WECOM_CORP_SECRET")),
		BaseURL:       strings.TrimSpace(os.Getenv("WECOM_BASE_URL")),
		TokenCache:    strings.TrimSpace(os.Getenv("WECOM_TOKEN_CACHE")),
		ResourceTable: strings.TrimSpace(os.Getenv("WECOM_RESOURCE_TABLE")),
		HTTPClient:    &http.Client{Timeout: defaultHTTPTimeout},
	}

	fs := flag.NewFlagSet("wecom-cli", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&cfg.CorpID, "corpid", cfg.CorpID, "WeCom enterprise ID")
	fs.StringVar(&cfg.CorpSecret, "corpsecret", cfg.CorpSecret, "WeCom app secret")
	fs.StringVar(&cfg.BaseURL, "base-url", cfg.BaseURL, "WeCom API base URL")
	fs.StringVar(&cfg.TokenCache, "token-cache", cfg.TokenCache, "access_token cache file")
	fs.StringVar(&cfg.ResourceTable, "resource-table", cfg.ResourceTable, "created resource table file")
	if err := fs.Parse(args); err != nil {
		return cfg, nil, err
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	if cfg.TokenCache == "" {
		cfg.TokenCache = filepath.Join(homeDir(), ".wecom-cli", "access_tokens.json")
	}
	if cfg.ResourceTable == "" {
		cfg.ResourceTable = filepath.Join(homeDir(), ".wecom-cli", "resources.json")
	}
	return cfg, fs.Args(), nil
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("%s:%d: expected KEY=value", path, lineNum)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("%s:%d: empty environment variable name", path, lineNum)
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		value = strings.TrimSpace(value)
		if unquoted, err := strconv.Unquote(value); err == nil {
			value = unquoted
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("%s:%d: set %s: %w", path, lineNum, key, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	return nil
}

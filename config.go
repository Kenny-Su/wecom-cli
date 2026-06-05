package main

import (
	"bufio"
	"encoding/json"
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
	defaultHTTPTimeout = 30 * time.Second
)

type config struct {
	GatewayBaseURL string
	AGWBaseURL     string
	GatewayToken   string
	IdentityFile   string
	HTTPClient     *http.Client
}

func parseGlobalFlags(args []string) (config, []string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return config{}, nil, fmt.Errorf("resolve executable path: %w", err)
	}
	systemIdentityFile := strings.TrimSpace(os.Getenv("CLI_IDENTITY_FILE"))
	if err := loadDotEnvForExecutable(executablePath); err != nil {
		return config{}, nil, err
	}

	cfg := config{
		GatewayBaseURL: strings.TrimSpace(os.Getenv("WECOM_GATEWAY_BASE_URL")),
		AGWBaseURL:     strings.TrimSpace(os.Getenv("AGW_GATEWAY_BASE_URL")),
		IdentityFile:   systemIdentityFile,
		HTTPClient:     &http.Client{Timeout: defaultHTTPTimeout},
	}

	fs := flag.NewFlagSet("wecom-cli", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&cfg.GatewayBaseURL, "gateway-base-url", cfg.GatewayBaseURL, "WeCom relay gateway base URL")
	fs.StringVar(&cfg.AGWBaseURL, "agw-base-url", cfg.AGWBaseURL, "AGW admin backend gateway base URL")
	if err := fs.Parse(args); err != nil {
		return cfg, nil, err
	}
	cfg.GatewayBaseURL = strings.TrimRight(cfg.GatewayBaseURL, "/")
	cfg.AGWBaseURL = strings.TrimRight(firstNonBlank(cfg.AGWBaseURL, deriveAGWBaseURL(cfg.GatewayBaseURL)), "/")
	cfg.IdentityFile = strings.TrimSpace(cfg.IdentityFile)
	return cfg, fs.Args(), nil
}

func deriveAGWBaseURL(gatewayBaseURL string) string {
	base := strings.TrimRight(strings.TrimSpace(gatewayBaseURL), "/")
	if strings.HasSuffix(base, "/wecom") {
		return strings.TrimSuffix(base, "/wecom")
	}
	return base
}

func loadDotEnvForExecutable(executablePath string) error {
	if err := loadDotEnv(".env"); err != nil {
		return err
	}
	skillEnv := findSkillDotEnv(executablePath)
	if skillEnv == "" {
		return nil
	}
	return loadDotEnv(skillEnv)
}

func findSkillDotEnv(executablePath string) string {
	dir := filepath.Dir(executablePath)
	for {
		if fileExists(filepath.Join(dir, "SKILL.md")) {
			return filepath.Join(dir, ".env")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
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
		if key == "CLI_IDENTITY_FILE" {
			continue
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

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func readAccessToken(identityFile string) (string, error) {
	values, err := readIdentityFile(identityFile)
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(values["ACCESS_TOKEN"])
	if token == "" {
		return "", fmt.Errorf("%s: ACCESS_TOKEN is required", identityFile)
	}
	return token, nil
}

func readIdentityFile(path string) (map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	values := map[string]string{}
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, fmt.Errorf("parse identity JSON %s: %w", path, err)
	}
	return values, nil
}

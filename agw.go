package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func resolveUsers(cfg config, userIDs []string, names []string) ([]string, error) {
	resolved := make([]string, 0, len(userIDs)+len(names))
	for _, userID := range userIDs {
		userID = strings.TrimSpace(userID)
		if userID != "" {
			resolved = append(resolved, userID)
		}
	}
	for _, name := range names {
		userID, err := resolveUserName(cfg, name)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, userID)
	}
	if len(resolved) > 1000 {
		return nil, errors.New("user list can include at most 1000 users")
	}
	return resolved, nil
}

func resolveUserName(cfg config, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("empty employee name")
	}
	if cfg.AGWCLI == "" {
		return "", errors.New("name lookup requires --agw-cli or AGW_CLI")
	}
	if strings.TrimSpace(cfg.CorpID) == "" {
		return "", errors.New("name lookup requires --corpid or WECOM_CORP_ID")
	}
	cmd := exec.Command(cfg.AGWCLI, "employee-wecom-mapping", "get-by-user-name", "--corpid", cfg.CorpID, "--user-name", name)
	cmd.Env = os.Environ()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	raw, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("resolve %q with agw-cli: %w: %s", name, err, strings.TrimSpace(stderr.String()))
	}
	userID, err := extractQWUserID(raw)
	if err != nil {
		return "", fmt.Errorf("resolve %q with agw-cli: %w", name, err)
	}
	return userID, nil
}

func extractQWUserID(raw []byte) (string, error) {
	var response struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return "", fmt.Errorf("parse lookup JSON: %w", err)
	}
	if response.Code != 0 {
		return "", fmt.Errorf("lookup returned code %d: %s", response.Code, response.Msg)
	}
	var data struct {
		QWUserID string `json:"qwUserid"`
	}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		return "", fmt.Errorf("parse lookup data: %w", err)
	}
	if strings.TrimSpace(data.QWUserID) == "" {
		return "", errors.New("lookup response did not include data.qwUserid")
	}
	return strings.TrimSpace(data.QWUserID), nil
}

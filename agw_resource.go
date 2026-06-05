package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type resourceTrackSpec struct {
	ResourceType  string
	PlatformField string
	IDFields      []string
	Name          string
	Command       string
	Request       any
}

type agwResourceRequest struct {
	ID            int64  `json:"id,omitempty"`
	ResourceType  string `json:"resourceType"`
	PlatformField string `json:"platformField"`
	ExternalID    string `json:"externalId"`
	Name          string `json:"name,omitempty"`
	ParentID      *int64 `json:"parentId,omitempty"`
	MetadataJSON  string `json:"metadataJson,omitempty"`
	ChatID        string `json:"chatId,omitempty"`
	RunID         string `json:"runId,omitempty"`
	Status        string `json:"status,omitempty"`
}

func runResources(c *wecomClient, args []string) error {
	if len(args) == 0 || isHelp(args[0]) {
		printResourcesUsage()
		return nil
	}
	switch args[0] {
	case "add":
		return resourceAdd(c, args[1:])
	case "get":
		return resourceGet(c, args[1:])
	case "list":
		return resourceList(c, args[1:])
	case "update":
		return resourceUpdate(c, args[1:])
	case "delete":
		return resourceDelete(c, args[1:])
	default:
		return fmt.Errorf("unknown resources command %q", args[0])
	}
}

func resourceAdd(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("resources add", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	req, parentID := bindResourceFlags(fs, false)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *parentID > 0 {
		req.ParentID = parentID
	}
	if err := requireResourceIdentity(req); err != nil {
		return err
	}
	return c.postAGWJSON("/adm/api/user-agent-resource/add", req)
}

func resourceGet(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("resources get", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.Int64("id", 0, "resource ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id <= 0 {
		return errors.New("--id must be greater than 0")
	}
	return c.getAGWJSON(fmt.Sprintf("/adm/api/service/user-agent-resource/%d", *id), nil)
}

func resourceList(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("resources list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	pageNum := fs.Int("page-num", 0, "page number")
	pageSize := fs.Int("page-size", 100, "page size")
	resourceType := fs.String("resource-type", "", "resource type")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *pageNum < 0 {
		return errors.New("--page-num must be greater than or equal to 0")
	}
	if *pageSize <= 0 {
		return errors.New("--page-size must be greater than 0")
	}
	q := url.Values{}
	q.Set("pageNum", strconv.Itoa(*pageNum))
	q.Set("pageSize", strconv.Itoa(*pageSize))
	addQuery(q, "resourceType", *resourceType)
	return c.getAGWJSON("/adm/api/service/user-agent-resources", q)
}

func resourceUpdate(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("resources update", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	req, parentID := bindResourceFlags(fs, true)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if req.ID <= 0 {
		return errors.New("--id must be greater than 0")
	}
	if *parentID > 0 {
		req.ParentID = parentID
	}
	if err := requireResourceIdentity(req); err != nil {
		return err
	}
	return c.postAGWJSON(fmt.Sprintf("/adm/api/user-agent-resource/update/%d", req.ID), req)
}

func resourceDelete(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("resources delete", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.Int64("id", 0, "resource ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id <= 0 {
		return errors.New("--id must be greater than 0")
	}
	return c.postAGWJSON(fmt.Sprintf("/adm/api/user-agent-resource/delete/%d", *id), nil)
}

func bindResourceFlags(fs *flag.FlagSet, includeID bool) (*agwResourceRequest, *int64) {
	req := &agwResourceRequest{}
	if includeID {
		fs.Int64Var(&req.ID, "id", 0, "resource ID")
	}
	fs.StringVar(&req.ResourceType, "resource-type", "", "resource type")
	fs.StringVar(&req.PlatformField, "platform-field", "", "platform field")
	fs.StringVar(&req.ExternalID, "external-id", "", "external resource ID")
	fs.StringVar(&req.Name, "name", "", "display name")
	parentID := fs.Int64("parent-id", 0, "parent resource ID")
	fs.StringVar(&req.MetadataJSON, "metadata-json", "", "metadata JSON string")
	fs.StringVar(&req.ChatID, "chat-id", "", "chat ID")
	fs.StringVar(&req.RunID, "run-id", "", "run ID")
	fs.StringVar(&req.Status, "status", "", "status")
	return req, parentID
}

func requireResourceIdentity(req *agwResourceRequest) error {
	missing := requiredMissing(map[string]string{
		"--resource-type":  req.ResourceType,
		"--platform-field": req.PlatformField,
		"--external-id":    req.ExternalID,
	})
	if len(missing) > 0 {
		return fmt.Errorf("missing required flags: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (c *wecomClient) postWeComAndStore(path string, body any, spec resourceTrackSpec) error {
	raw, err := c.postWeComRaw(path, body)
	if err != nil {
		return err
	}
	if err := printRawResponse(raw); err != nil {
		return err
	}
	return c.storeCreatedResource(spec, raw)
}

func (c *wecomClient) storeCreatedResource(spec resourceTrackSpec, raw []byte) error {
	response := map[string]any{}
	if len(bytes.TrimSpace(raw)) == 0 || !json.Valid(raw) {
		return nil
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil
	}
	externalID := firstResourceID(response, spec.IDFields)
	if externalID == "" {
		return nil
	}
	req := agwResourceRequest{
		ResourceType:  spec.ResourceType,
		PlatformField: spec.PlatformField,
		ExternalID:    externalID,
		Name:          spec.Name,
		Status:        "ACTIVE",
		MetadataJSON:  resourceMetadataJSON(spec, response),
	}
	_, err := c.doAGWRaw(http.MethodPost, "/adm/api/user-agent-resource/add", req)
	return err
}

func resourceMetadataJSON(spec resourceTrackSpec, response map[string]any) string {
	metadata := map[string]any{
		"command":  spec.Command,
		"response": response,
		"storedAt": time.Now().UTC().Format(time.RFC3339),
	}
	if request := mapFromAny(spec.Request); len(request) > 0 {
		metadata["request"] = request
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return ""
	}
	return string(raw)
}

func firstResourceID(value any, fields []string) string {
	for _, field := range fields {
		if found := findStringField(value, field); found != "" {
			return found
		}
	}
	return ""
}

func findStringField(value any, field string) string {
	switch typed := value.(type) {
	case map[string]any:
		if raw, ok := typed[field]; ok {
			if str, ok := raw.(string); ok && strings.TrimSpace(str) != "" {
				return strings.TrimSpace(str)
			}
		}
		for _, nested := range typed {
			if found := findStringField(nested, field); found != "" {
				return found
			}
		}
	case []any:
		for _, nested := range typed {
			if found := findStringField(nested, field); found != "" {
				return found
			}
		}
	}
	return ""
}

func mapFromAny(value any) map[string]any {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	redactTrackedValue(out)
	return out
}

func redactTrackedValue(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			switch key {
			case "file_base64_content", "password", "selected_ticket":
				typed[key] = "[redacted]"
			default:
				redactTrackedValue(nested)
			}
		}
	case []any:
		for _, nested := range typed {
			redactTrackedValue(nested)
		}
	}
}

func (c *wecomClient) getAGWJSON(path string, query url.Values) error {
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	raw, err := c.doAGWRaw(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return printRawResponse(raw)
}

func (c *wecomClient) postAGWJSON(path string, body any) error {
	raw, err := c.doAGWRaw(http.MethodPost, path, body)
	if err != nil {
		return err
	}
	return printRawResponse(raw)
}

func (c *wecomClient) doAGWRaw(method string, path string, body any) ([]byte, error) {
	if err := c.requireGatewayToken(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(c.cfg.AGWBaseURL) == "" {
		return nil, errors.New("AGW backend URL is required; set --agw-base-url, AGW_GATEWAY_BASE_URL, or WECOM_AGW_BASE_URL")
	}

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, c.cfg.AGWBaseURL+path, reader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.GatewayToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
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
		return nil, fmt.Errorf("%s %s returned HTTP %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return raw, nil
}

func addQuery(q url.Values, key string, value string) {
	if strings.TrimSpace(value) != "" {
		q.Set(key, value)
	}
}

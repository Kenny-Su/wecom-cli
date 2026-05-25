package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type resourceTrackSpec struct {
	Type     string
	IDFields []string
	Name     string
	Command  string
	Request  any
}

type resourceTable struct {
	Resources []resourceRecord `json:"resources"`
}

type resourceRecord struct {
	Type      string         `json:"type"`
	ID        string         `json:"id"`
	Name      string         `json:"name,omitempty"`
	Command   string         `json:"command"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at,omitempty"`
	Request   map[string]any `json:"request,omitempty"`
	Response  map[string]any `json:"response,omitempty"`
}

func runResources(cfg config, args []string) error {
	if len(args) == 0 || isHelp(args[0]) {
		printResourcesUsage()
		return nil
	}
	switch args[0] {
	case "list":
		return resourcesList(cfg, args[1:])
	case "path":
		fmt.Println(expandPath(cfg.ResourceTable))
		return nil
	default:
		return fmt.Errorf("unknown resources command %q", args[0])
	}
}

func resourcesList(cfg config, args []string) error {
	fs := flag.NewFlagSet("resources list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	resourceType := fs.String("type", "", "filter by resource type")
	jsonOut := fs.Bool("json", false, "print the raw resource table as JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	table, err := readResourceTable(cfg.ResourceTable)
	if err != nil {
		return err
	}
	if *resourceType != "" {
		filtered := table.Resources[:0]
		for _, rec := range table.Resources {
			if rec.Type == *resourceType {
				filtered = append(filtered, rec)
			}
		}
		table.Resources = filtered
	}
	sort.SliceStable(table.Resources, func(i, j int) bool {
		return table.Resources[i].CreatedAt < table.Resources[j].CreatedAt
	})
	if *jsonOut {
		return printPrettyJSON(table)
	}
	printResourceRecords(table.Resources)
	return nil
}

func trackCreatedResource(c *wecomClient, spec resourceTrackSpec, response map[string]any) error {
	id := firstResourceID(response, spec.IDFields)
	if id == "" {
		return nil
	}
	request := mapFromAny(spec.Request)
	record := resourceRecord{
		Type:      spec.Type,
		ID:        id,
		Name:      spec.Name,
		Command:   spec.Command,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Request:   request,
		Response:  response,
	}
	return upsertResourceRecord(c.cfg.ResourceTable, record)
}

func upsertResourceRecord(path string, record resourceRecord) error {
	path = expandPath(path)
	table, err := readResourceTable(path)
	if err != nil {
		return err
	}
	for i, existing := range table.Resources {
		if existing.Type == record.Type && existing.ID == record.ID {
			record.CreatedAt = existing.CreatedAt
			record.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			table.Resources[i] = record
			return writeResourceTable(path, table)
		}
	}
	table.Resources = append(table.Resources, record)
	return writeResourceTable(path, table)
}

func readResourceTable(path string) (resourceTable, error) {
	raw, err := os.ReadFile(expandPath(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return resourceTable{}, nil
		}
		return resourceTable{}, fmt.Errorf("read resource table: %w", err)
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return resourceTable{}, nil
	}
	var table resourceTable
	if err := json.Unmarshal(raw, &table); err != nil {
		return resourceTable{}, fmt.Errorf("parse resource table: %w", err)
	}
	return table, nil
}

func writeResourceTable(path string, table resourceTable) error {
	raw, err := json.MarshalIndent(table, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal resource table: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create resource table directory: %w", err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o600); err != nil {
		return fmt.Errorf("write resource table: %w", err)
	}
	return nil
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

func printResourceRecords(records []resourceRecord) {
	if len(records) == 0 {
		fmt.Println("No resources tracked.")
		return
	}
	fmt.Printf("%-18s  %-34s  %-28s  %s\n", "TYPE", "ID", "NAME", "CREATED")
	for _, rec := range records {
		fmt.Printf("%-18s  %-34s  %-28s  %s\n", rec.Type, rec.ID, rec.Name, rec.CreatedAt)
	}
}

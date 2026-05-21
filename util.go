package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type stringList []string

func (s *stringList) String() string {
	return strings.Join(*s, ",")
}

func (s *stringList) Set(value string) error {
	value = strings.TrimSpace(value)
	if value != "" {
		*s = append(*s, value)
	}
	return nil
}

type intList []int

func (s *intList) String() string {
	values := make([]string, 0, len(*s))
	for _, value := range *s {
		values = append(values, fmt.Sprintf("%d", value))
	}
	return strings.Join(values, ",")
}

func (s *intList) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	*s = append(*s, parsed)
	return nil
}

func expandPath(path string) string {
	if path == "~" {
		return homeDir()
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir(), strings.TrimPrefix(path, "~/"))
	}
	return path
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "."
	}
	return home
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		out = append(out, value)
		seen[value] = true
	}
	return out
}

func parseUserIDs(userIDs []string) ([]string, error) {
	users := uniqueStrings(userIDs)
	if len(users) > 1000 {
		return nil, fmt.Errorf("user list can include at most 1000 users")
	}
	return users, nil
}

func uniqueShares(values []calendarShare) []calendarShare {
	seen := map[string]bool{}
	var out []calendarShare
	for _, value := range values {
		value.UserID = strings.TrimSpace(value.UserID)
		if value.UserID == "" || seen[value.UserID] {
			continue
		}
		out = append(out, value)
		seen[value.UserID] = true
	}
	return out
}

func requiredMissing(values map[string]string) []string {
	var missing []string
	for name, value := range values {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	return missing
}

func printPrettyJSON(value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(raw))
	return nil
}

func isHelp(arg string) bool {
	return arg == "help" || arg == "-h" || arg == "--help"
}

func intPtr(value int) *int {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

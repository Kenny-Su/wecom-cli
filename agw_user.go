package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
)

func runUsers(c *wecomClient, args []string) error {
	if len(args) == 0 || isHelp(args[0]) {
		printUsersUsage()
		return nil
	}
	switch args[0] {
	case "get":
		return userGet(c, args[1:])
	case "get-by-qw-user":
		return userLookup(c, args[1:], "/adm/api/service/employee-wecom-mapping/by-qw-user", "qwUserid", "--qw-userid")
	case "get-by-staff-id":
		return userLookup(c, args[1:], "/adm/api/service/employee-wecom-mapping/by-staff-id", "staffId", "--staff-id")
	case "get-by-legacy-staff-id":
		return userLookup(c, args[1:], "/adm/api/service/employee-wecom-mapping/by-legacy-staff-id", "legacyStaffId", "--legacy-staff-id")
	case "get-by-name", "get-by-user-name":
		return userLookup(c, args[1:], "/adm/api/service/employee-wecom-mapping/by-user-name", "userName", "--user-name")
	case "list":
		return userList(c, args[1:])
	default:
		return fmt.Errorf("unknown users command %q", args[0])
	}
}

func userGet(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("users get", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.Int64("id", 0, "mapping ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id <= 0 {
		return errors.New("--id must be greater than 0")
	}
	return c.getAGWJSON(fmt.Sprintf("/adm/api/service/employee-wecom-mapping/%d", *id), nil)
}

func userLookup(c *wecomClient, args []string, path string, queryKey string, flagName string) error {
	fs := flag.NewFlagSet("users lookup", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	value := fs.String(strings.TrimPrefix(flagName, "--"), "", flagName)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if missing := requiredMissing(map[string]string{flagName: *value}); len(missing) > 0 {
		return fmt.Errorf("missing required flags: %s", strings.Join(missing, ", "))
	}
	q := url.Values{}
	q.Set(queryKey, *value)
	return c.getAGWJSON(path, q)
}

func userList(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("users list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	pageNum := fs.Int("page-num", 0, "page number")
	pageSize := fs.Int("page-size", 100, "page size")
	qwUserid := fs.String("qw-userid", "", "WeCom user ID")
	staffID := fs.String("staff-id", "", "staff ID")
	legacyStaffID := fs.String("legacy-staff-id", "", "legacy staff ID")
	mobile := fs.String("mobile", "", "mobile number")
	userName := fs.String("user-name", "", "user name")
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
	addQuery(q, "qwUserid", *qwUserid)
	addQuery(q, "staffId", *staffID)
	addQuery(q, "legacyStaffId", *legacyStaffID)
	addQuery(q, "mobile", *mobile)
	addQuery(q, "userName", *userName)
	return c.getAGWJSON("/adm/api/service/employee-wecom-mappings", q)
}

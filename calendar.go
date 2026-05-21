package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type calendarAddRequest struct {
	Calendar calendarPayload `json:"calendar"`
	AgentID  int64           `json:"agentid,omitempty"`
}

type calendarUpdateRequest struct {
	SkipPublicRange int             `json:"skip_public_range,omitempty"`
	Calendar        calendarPayload `json:"calendar"`
}

type calendarGetRequest struct {
	CalIDList []string `json:"cal_id_list"`
}

type calendarDeleteRequest struct {
	CalID string `json:"cal_id"`
}

type calendarPayload struct {
	CalID          string          `json:"cal_id,omitempty"`
	Admins         []string        `json:"admins,omitempty"`
	SetAsDefault   int             `json:"set_as_default,omitempty"`
	Summary        string          `json:"summary"`
	Color          string          `json:"color,omitempty"`
	Description    string          `json:"description,omitempty"`
	Shares         []calendarShare `json:"shares,omitempty"`
	IsPublic       int             `json:"is_public,omitempty"`
	PublicRange    *publicRange    `json:"public_range,omitempty"`
	IsCorpCalendar int             `json:"is_corp_calendar,omitempty"`
}

type calendarShare struct {
	UserID     string `json:"userid"`
	Permission int    `json:"permission,omitempty"`
}

type publicRange struct {
	UserIDs  []string `json:"userids,omitempty"`
	PartyIDs []int64  `json:"partyids,omitempty"`
}

type calendarCreateInput struct {
	Summary        string
	Color          string
	Description    string
	SetAsDefault   bool
	IsPublic       bool
	IsCorpCalendar bool
	Admins         []string
	Shares         []string
	PublicUsers    []string
	PublicParties  []string
	AgentID        int64
}

type calendarUpdateInput struct {
	CalID           string
	Summary         string
	Color           string
	Description     string
	SkipPublicRange bool
	Admins          []string
	Shares          []string
	PublicUsers     []string
	PublicParties   []string
}

func runCalendar(c *wecomClient, args []string) error {
	switch args[0] {
	case "create":
		return calendarCreate(c, args[1:])
	case "update":
		return calendarUpdate(c, args[1:])
	case "get":
		return calendarGet(c, args[1:])
	case "delete":
		return calendarDelete(c, args[1:])
	case "help", "-h", "--help":
		printCalendarUsage()
		return nil
	default:
		return fmt.Errorf("unknown calendar command %q", args[0])
	}
}

func calendarCreate(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("calendar create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var admins, shares, publicUsers, publicParties stringList
	summary := fs.String("summary", "", "calendar title")
	color := fs.String("color", "", "calendar RGB color")
	description := fs.String("description", "", "calendar description")
	setDefault := fs.Bool("default", false, "set as app default calendar")
	isPublic := fs.Bool("public", false, "create public calendar")
	isCorp := fs.Bool("corp-calendar", false, "create all-staff calendar")
	agentID := fs.Int64("agentid", 0, "authorized app agentid")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	fs.Var(&admins, "admin", "admin userid; repeatable")
	fs.Var(&shares, "share", "userid[:permission]; repeatable")
	fs.Var(&publicUsers, "public-user", "public-range userid; repeatable")
	fs.Var(&publicParties, "public-party", "public-range department ID; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}

	req, err := buildCalendarCreateRequest(calendarCreateInput{
		Summary:        *summary,
		Color:          *color,
		Description:    *description,
		SetAsDefault:   *setDefault,
		IsPublic:       *isPublic,
		IsCorpCalendar: *isCorp,
		Admins:         admins,
		Shares:         shares,
		PublicUsers:    publicUsers,
		PublicParties:  publicParties,
		AgentID:        *agentID,
	})
	if err != nil {
		return err
	}
	if *dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	return c.addCalendar(req)
}

func calendarUpdate(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("calendar update", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var admins, shares, publicUsers, publicParties stringList
	calID := fs.String("cal-id", "", "calendar ID")
	summary := fs.String("summary", "", "calendar title")
	color := fs.String("color", "", "calendar RGB color")
	description := fs.String("description", "", "calendar description")
	skipPublicRange := fs.Bool("skip-public-range", false, "do not update public subscription range")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	fs.Var(&admins, "admin", "admin userid; repeatable")
	fs.Var(&shares, "share", "userid[:permission]; repeatable")
	fs.Var(&publicUsers, "public-user", "public-range userid; repeatable")
	fs.Var(&publicParties, "public-party", "public-range department ID; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}

	req, err := buildCalendarUpdateRequest(calendarUpdateInput{
		CalID:           *calID,
		Summary:         *summary,
		Color:           *color,
		Description:     *description,
		SkipPublicRange: *skipPublicRange,
		Admins:          admins,
		Shares:          shares,
		PublicUsers:     publicUsers,
		PublicParties:   publicParties,
	})
	if err != nil {
		return err
	}
	if *dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	return c.updateCalendar(req)
}

func calendarGet(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("calendar get", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var calIDs stringList
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	fs.Var(&calIDs, "cal-id", "calendar ID; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}

	req, err := buildCalendarGetRequest(calIDs)
	if err != nil {
		return err
	}
	if *dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	return c.getCalendar(req)
}

func calendarDelete(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("calendar delete", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	calID := fs.String("cal-id", "", "calendar ID")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}

	req, err := buildCalendarDeleteRequest(*calID)
	if err != nil {
		return err
	}
	if *dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	return c.deleteCalendar(req)
}

func buildCalendarCreateRequest(input calendarCreateInput) (calendarAddRequest, error) {
	summary, err := validateCalendarSummary(input.Summary)
	if err != nil {
		return calendarAddRequest{}, err
	}
	description, err := validateCalendarDescription(input.Description)
	if err != nil {
		return calendarAddRequest{}, err
	}
	color := strings.TrimSpace(input.Color)
	if !input.IsCorpCalendar {
		color, err = validateCalendarColor(color)
		if err != nil {
			if strings.Contains(err.Error(), "--color is required") {
				return calendarAddRequest{}, errors.New("--color is required unless --corp-calendar is set")
			}
			return calendarAddRequest{}, err
		}
	} else if color != "" {
		return calendarAddRequest{}, errors.New("--corp-calendar cannot be combined with --color")
	}
	if input.IsCorpCalendar && input.SetAsDefault {
		return calendarAddRequest{}, errors.New("--corp-calendar cannot be combined with --default")
	}

	adminIDs, err := parseUserIDs(input.Admins)
	if err != nil {
		return calendarAddRequest{}, err
	}
	if len(adminIDs) > 3 {
		return calendarAddRequest{}, errors.New("calendar admins can include at most 3 users")
	}
	shareIDs, err := parseShares(input.Shares)
	if err != nil {
		return calendarAddRequest{}, err
	}
	publicUserIDs, err := parseUserIDs(input.PublicUsers)
	if err != nil {
		return calendarAddRequest{}, err
	}
	partyIDs, err := parsePartyIDs(input.PublicParties)
	if err != nil {
		return calendarAddRequest{}, err
	}
	if input.IsCorpCalendar && !input.IsPublic {
		input.IsPublic = true
	}
	if input.IsCorpCalendar && len(publicUserIDs) == 0 && len(partyIDs) == 0 {
		return calendarAddRequest{}, errors.New("--corp-calendar requires --public-user or --public-party")
	}
	if !input.IsPublic && (len(publicUserIDs) > 0 || len(partyIDs) > 0) {
		return calendarAddRequest{}, errors.New("--public is required when public range flags are used")
	}

	payload := calendarPayload{
		Admins:      uniqueStrings(adminIDs),
		Summary:     summary,
		Color:       strings.ToUpper(color),
		Description: description,
		Shares:      shareIDs,
	}
	if input.SetAsDefault {
		payload.SetAsDefault = 1
	}
	if input.IsPublic {
		payload.IsPublic = 1
		payload.PublicRange = &publicRange{
			UserIDs:  uniqueStrings(publicUserIDs),
			PartyIDs: partyIDs,
		}
	}
	if input.IsCorpCalendar {
		payload.IsCorpCalendar = 1
	}
	return calendarAddRequest{Calendar: payload, AgentID: input.AgentID}, nil
}

func buildCalendarUpdateRequest(input calendarUpdateInput) (calendarUpdateRequest, error) {
	calID := strings.TrimSpace(input.CalID)
	if calID == "" {
		return calendarUpdateRequest{}, errors.New("--cal-id is required")
	}
	summary, err := validateCalendarSummary(input.Summary)
	if err != nil {
		return calendarUpdateRequest{}, err
	}
	description, err := validateCalendarDescription(input.Description)
	if err != nil {
		return calendarUpdateRequest{}, err
	}
	color, err := validateCalendarColor(input.Color)
	if err != nil {
		return calendarUpdateRequest{}, err
	}

	adminIDs, err := parseUserIDs(input.Admins)
	if err != nil {
		return calendarUpdateRequest{}, err
	}
	if len(adminIDs) > 3 {
		return calendarUpdateRequest{}, errors.New("calendar admins can include at most 3 users")
	}
	shareIDs, err := parseShares(input.Shares)
	if err != nil {
		return calendarUpdateRequest{}, err
	}
	publicUserIDs, err := parseUserIDs(input.PublicUsers)
	if err != nil {
		return calendarUpdateRequest{}, err
	}
	partyIDs, err := parsePartyIDs(input.PublicParties)
	if err != nil {
		return calendarUpdateRequest{}, err
	}
	if input.SkipPublicRange && (len(publicUserIDs) > 0 || len(partyIDs) > 0) {
		return calendarUpdateRequest{}, errors.New("--skip-public-range cannot be combined with public range flags")
	}

	payload := calendarPayload{
		CalID:       calID,
		Admins:      uniqueStrings(adminIDs),
		Summary:     summary,
		Color:       color,
		Description: description,
		Shares:      shareIDs,
	}
	if !input.SkipPublicRange && (len(publicUserIDs) > 0 || len(partyIDs) > 0) {
		payload.PublicRange = &publicRange{
			UserIDs:  uniqueStrings(publicUserIDs),
			PartyIDs: partyIDs,
		}
	}
	req := calendarUpdateRequest{Calendar: payload}
	if input.SkipPublicRange {
		req.SkipPublicRange = 1
	}
	return req, nil
}

func buildCalendarGetRequest(calIDs []string) (calendarGetRequest, error) {
	ids := uniqueStrings(calIDs)
	if len(ids) == 0 {
		return calendarGetRequest{}, errors.New("--cal-id is required")
	}
	if len(ids) > 1000 {
		return calendarGetRequest{}, errors.New("calendar get can include at most 1000 calendar IDs")
	}
	return calendarGetRequest{CalIDList: ids}, nil
}

func buildCalendarDeleteRequest(calID string) (calendarDeleteRequest, error) {
	calID = strings.TrimSpace(calID)
	if calID == "" {
		return calendarDeleteRequest{}, errors.New("--cal-id is required")
	}
	return calendarDeleteRequest{CalID: calID}, nil
}

func validateCalendarSummary(value string) (string, error) {
	summary := strings.TrimSpace(value)
	if summary == "" {
		return "", errors.New("--summary is required")
	}
	if len([]rune(summary)) > 128 {
		return "", errors.New("--summary must be 1 to 128 characters")
	}
	return summary, nil
}

func validateCalendarDescription(value string) (string, error) {
	description := strings.TrimSpace(value)
	if len([]rune(description)) > 512 {
		return "", errors.New("--description must be at most 512 characters")
	}
	return description, nil
}

func validateCalendarColor(value string) (string, error) {
	color := strings.TrimSpace(value)
	if color == "" {
		return "", errors.New("--color is required")
	}
	if !colorPattern.MatchString(color) {
		return "", errors.New("--color must be an RGB value like #2F7DFF")
	}
	return strings.ToUpper(color), nil
}

func parseShares(rawShares []string) ([]calendarShare, error) {
	var shares []calendarShare
	for _, raw := range rawShares {
		userID, permission, err := parseShareSpec(raw)
		if err != nil {
			return nil, err
		}
		shares = append(shares, calendarShare{UserID: userID, Permission: permission})
	}
	if len(shares) > 2000 {
		return nil, errors.New("calendar shares can include at most 2000 users")
	}
	return uniqueShares(shares), nil
}

func parseShareSpec(raw string) (string, int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", 0, errors.New("empty share value")
	}
	userID, permissionText, found := strings.Cut(value, ":")
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", 0, fmt.Errorf("invalid share %q: missing userid", raw)
	}
	if !found || strings.TrimSpace(permissionText) == "" {
		return userID, 1, nil
	}
	permission, err := strconv.Atoi(strings.TrimSpace(permissionText))
	if err != nil || (permission != 1 && permission != 3) {
		return "", 0, fmt.Errorf("invalid share %q: permission must be 1 or 3", raw)
	}
	return userID, permission, nil
}

func parsePartyIDs(rawParties []string) ([]int64, error) {
	var ids []int64
	seen := map[int64]bool{}
	for _, raw := range rawParties {
		id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("invalid --public-party %q: must be a positive integer", raw)
		}
		if !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	if len(ids) > 100 {
		return nil, errors.New("public range can include at most 100 departments")
	}
	return ids, nil
}

func (c *wecomClient) addCalendar(req calendarAddRequest) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	path := "/cgi-bin/oa/calendar/add?access_token=" + url.QueryEscape(token)
	return c.postWeCom(path, req)
}

func (c *wecomClient) updateCalendar(req calendarUpdateRequest) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	path := "/cgi-bin/oa/calendar/update?access_token=" + url.QueryEscape(token)
	return c.postWeCom(path, req)
}

func (c *wecomClient) getCalendar(req calendarGetRequest) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	path := "/cgi-bin/oa/calendar/get?access_token=" + url.QueryEscape(token)
	return c.postWeCom(path, req)
}

func (c *wecomClient) deleteCalendar(req calendarDeleteRequest) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	path := "/cgi-bin/oa/calendar/del?access_token=" + url.QueryEscape(token)
	return c.postWeCom(path, req)
}

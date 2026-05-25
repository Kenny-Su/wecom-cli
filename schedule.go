package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type scheduleRequest struct {
	Schedule schedulePayload `json:"schedule"`
	AgentID  int64           `json:"agentid,omitempty"`
}

type scheduleUpdateRequest struct {
	SkipAttendees int             `json:"skip_attendees,omitempty"`
	OpMode        int             `json:"op_mode,omitempty"`
	OpStartTime   int64           `json:"op_start_time,omitempty"`
	Schedule      schedulePayload `json:"schedule"`
}

type scheduleGetRequest struct {
	ScheduleIDList []string `json:"schedule_id_list"`
}

type scheduleListRequest struct {
	CalID  string `json:"cal_id"`
	Offset int    `json:"offset,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

type scheduleDeleteRequest struct {
	ScheduleID  string `json:"schedule_id"`
	OpMode      int    `json:"op_mode,omitempty"`
	OpStartTime int64  `json:"op_start_time,omitempty"`
}

type scheduleAttendeesRequest struct {
	ScheduleID string             `json:"schedule_id"`
	Attendees  []scheduleAttendee `json:"attendees,omitempty"`
}

type schedulePayload struct {
	ScheduleID  string             `json:"schedule_id,omitempty"`
	Admins      []string           `json:"admins,omitempty"`
	StartTime   int64              `json:"start_time,omitempty"`
	EndTime     int64              `json:"end_time,omitempty"`
	IsWholeDay  int                `json:"is_whole_day,omitempty"`
	Attendees   []scheduleAttendee `json:"attendees,omitempty"`
	Summary     string             `json:"summary,omitempty"`
	Description string             `json:"description,omitempty"`
	Reminders   *scheduleReminder  `json:"reminders,omitempty"`
	Location    string             `json:"location,omitempty"`
	CalID       string             `json:"cal_id,omitempty"`
}

type scheduleAttendee struct {
	UserID string `json:"userid"`
}

type scheduleReminder struct {
	IsRemind              int   `json:"is_remind,omitempty"`
	RemindBeforeEventSecs *int  `json:"remind_before_event_secs,omitempty"`
	RemindTimeDiffs       []int `json:"remind_time_diffs,omitempty"`
	IsRepeat              int   `json:"is_repeat,omitempty"`
	RepeatType            *int  `json:"repeat_type,omitempty"`
	RepeatUntil           int64 `json:"repeat_until,omitempty"`
	IsCustomRepeat        int   `json:"is_custom_repeat,omitempty"`
	RepeatInterval        int   `json:"repeat_interval,omitempty"`
	RepeatDayOfWeek       []int `json:"repeat_day_of_week,omitempty"`
	RepeatDayOfMonth      []int `json:"repeat_day_of_month,omitempty"`
	Timezone              *int  `json:"timezone,omitempty"`
}

type scheduleInput struct {
	ScheduleID            string
	CalID                 string
	Summary               string
	Description           string
	Location              string
	Start                 string
	End                   string
	WholeDay              bool
	Admins                []string
	Attendees             []string
	Remind                bool
	RemindBeforeEventSecs int
	RemindTimeDiffs       []int
	Repeat                bool
	RepeatType            int
	RepeatUntil           string
	CustomRepeat          bool
	RepeatInterval        int
	RepeatDayOfWeek       []int
	RepeatDayOfMonth      []int
	Timezone              *int
	AgentID               int64
}

type scheduleUpdateInput struct {
	*scheduleInput
	SkipAttendees bool
	OpMode        int
	OpStartTime   string
}

func runSchedule(c *wecomClient, args []string) error {
	switch args[0] {
	case "create":
		return scheduleCreate(c, args[1:])
	case "update":
		return scheduleUpdate(c, args[1:])
	case "get":
		return scheduleGet(c, args[1:])
	case "list":
		return scheduleList(c, args[1:])
	case "delete", "cancel":
		return scheduleDelete(c, args[1:])
	case "add-attendees":
		return scheduleChangeAttendees(c, args[1:], true)
	case "remove-attendees":
		return scheduleChangeAttendees(c, args[1:], false)
	case "help", "-h", "--help":
		printScheduleUsage()
		return nil
	default:
		return fmt.Errorf("unknown schedule command %q", args[0])
	}
}

func scheduleCreate(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("schedule create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	input, dryRun := bindScheduleFlags(fs, false)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildScheduleCreateRequest(c.cfg, *input)
	if err != nil {
		return err
	}
	if *dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	return c.addSchedule(req)
}

func scheduleUpdate(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("schedule update", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	input, dryRun := bindScheduleUpdateFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildScheduleUpdateRequest(c.cfg, *input)
	if err != nil {
		return err
	}
	if *dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	return c.updateSchedule(req)
}

func scheduleGet(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("schedule get", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var ids stringList
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	fs.Var(&ids, "schedule-id", "schedule ID; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildScheduleGetRequest(ids)
	if err != nil {
		return err
	}
	if *dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	return c.getSchedule(req)
}

func scheduleList(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("schedule list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	calID := fs.String("cal-id", "", "calendar ID")
	offset := fs.Int("offset", 0, "pagination offset")
	limit := fs.Int("limit", 500, "pagination limit, 1-1000")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildScheduleListRequest(*calID, *offset, *limit)
	if err != nil {
		return err
	}
	if *dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	return c.listSchedule(req)
}

func scheduleDelete(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("schedule delete", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	scheduleID := fs.String("schedule-id", "", "schedule ID")
	opMode := fs.Int("op-mode", 0, "repeat operation mode: 0 all, 1 current, 2 future")
	opStartTime := fs.String("op-start-time", "", "repeat occurrence start time")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildScheduleDeleteRequest(*scheduleID, *opMode, *opStartTime)
	if err != nil {
		return err
	}
	if *dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	return c.deleteSchedule(req)
}

func scheduleChangeAttendees(c *wecomClient, args []string, add bool) error {
	name := "schedule remove-attendees"
	if add {
		name = "schedule add-attendees"
	}
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	scheduleID := fs.String("schedule-id", "", "schedule ID")
	var attendees stringList
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	fs.Var(&attendees, "attendee", "attendee userid; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildScheduleAttendeesRequest(*scheduleID, attendees)
	if err != nil {
		return err
	}
	if *dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	if add {
		return c.addScheduleAttendees(req)
	}
	return c.removeScheduleAttendees(req)
}

func bindScheduleFlags(fs *flag.FlagSet, includeID bool) (*scheduleInput, *bool) {
	input := &scheduleInput{}
	if includeID {
		fs.StringVar(&input.ScheduleID, "schedule-id", "", "schedule ID")
	}
	fs.StringVar(&input.CalID, "cal-id", "", "calendar ID")
	fs.StringVar(&input.Summary, "summary", "", "schedule title")
	fs.StringVar(&input.Description, "description", "", "schedule description")
	fs.StringVar(&input.Location, "location", "", "schedule location")
	fs.StringVar(&input.Start, "start", "", "start time, Unix seconds or RFC3339")
	fs.StringVar(&input.End, "end", "", "end time, Unix seconds or RFC3339")
	fs.BoolVar(&input.WholeDay, "whole-day", false, "mark schedule as whole day")
	if !includeID {
		fs.Int64Var(&input.AgentID, "agentid", 0, "authorized app agentid")
	}
	fs.Var((*stringList)(&input.Admins), "admin", "admin userid; repeatable")
	fs.Var((*stringList)(&input.Attendees), "attendee", "attendee userid; repeatable")
	fs.BoolVar(&input.Remind, "remind", false, "enable reminder")
	fs.IntVar(&input.RemindBeforeEventSecs, "remind-before", -1, "seconds before event reminder")
	fs.Var((*intList)(&input.RemindTimeDiffs), "remind-time-diff", "reminder diff from start time in seconds; repeatable")
	fs.BoolVar(&input.Repeat, "repeat", false, "enable repeat schedule")
	fs.IntVar(&input.RepeatType, "repeat-type", -1, "repeat type: 0 daily, 1 weekly, 2 monthly, 5 yearly, 7 workday")
	fs.StringVar(&input.RepeatUntil, "repeat-until", "", "repeat end time, Unix seconds or RFC3339")
	fs.BoolVar(&input.CustomRepeat, "custom-repeat", false, "enable custom repeat")
	fs.IntVar(&input.RepeatInterval, "repeat-interval", 0, "custom repeat interval")
	fs.Var((*intList)(&input.RepeatDayOfWeek), "repeat-day-of-week", "custom repeat weekday; repeatable")
	fs.Var((*intList)(&input.RepeatDayOfMonth), "repeat-day-of-month", "custom repeat month day; repeatable")
	timezone := fs.Int("timezone", 999, "UTC offset timezone, -12 to 12")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	fs.Func("timezone-set", "deprecated; use --timezone", func(value string) error {
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return err
		}
		*timezone = parsed
		return nil
	})
	input.Timezone = timezone
	return input, dryRun
}

func bindScheduleUpdateFlags(fs *flag.FlagSet) (*scheduleUpdateInput, *bool) {
	base, dryRun := bindScheduleFlags(fs, true)
	input := &scheduleUpdateInput{scheduleInput: base}
	fs.BoolVar(&input.SkipAttendees, "skip-attendees", false, "do not update attendees")
	fs.IntVar(&input.OpMode, "op-mode", 0, "repeat operation mode: 0 all, 1 current, 2 future")
	fs.StringVar(&input.OpStartTime, "op-start-time", "", "repeat occurrence start time")
	return input, dryRun
}

func buildScheduleCreateRequest(cfg config, input scheduleInput) (scheduleRequest, error) {
	payload, err := buildSchedulePayload(cfg, input, false)
	if err != nil {
		return scheduleRequest{}, err
	}
	return scheduleRequest{Schedule: payload, AgentID: input.AgentID}, nil
}

func buildScheduleUpdateRequest(cfg config, input scheduleUpdateInput) (scheduleUpdateRequest, error) {
	if strings.TrimSpace(input.ScheduleID) == "" {
		return scheduleUpdateRequest{}, errors.New("--schedule-id is required")
	}
	payload, err := buildSchedulePayload(cfg, *input.scheduleInput, true)
	if err != nil {
		return scheduleUpdateRequest{}, err
	}
	req := scheduleUpdateRequest{
		OpMode:   input.OpMode,
		Schedule: payload,
	}
	if input.SkipAttendees {
		req.SkipAttendees = 1
	}
	if input.OpMode != 0 {
		if input.OpMode != 1 && input.OpMode != 2 {
			return scheduleUpdateRequest{}, errors.New("--op-mode must be 0, 1, or 2")
		}
		opStart, err := parseOptionalTime(input.OpStartTime, "--op-start-time")
		if err != nil {
			return scheduleUpdateRequest{}, err
		}
		if opStart == 0 {
			return scheduleUpdateRequest{}, errors.New("--op-start-time is required when --op-mode is 1 or 2")
		}
		req.OpStartTime = opStart
	}
	return req, nil
}

func buildSchedulePayload(cfg config, input scheduleInput, update bool) (schedulePayload, error) {
	startTime, err := parseRequiredTime(input.Start, "--start")
	if err != nil {
		return schedulePayload{}, err
	}
	endTime, err := parseRequiredTime(input.End, "--end")
	if err != nil {
		return schedulePayload{}, err
	}
	if endTime <= startTime {
		return schedulePayload{}, errors.New("--end must be after --start")
	}
	if len([]rune(strings.TrimSpace(input.Summary))) > 128 {
		return schedulePayload{}, errors.New("--summary must be at most 128 characters")
	}
	if len([]rune(strings.TrimSpace(input.Description))) > 1000 {
		return schedulePayload{}, errors.New("--description must be at most 1000 characters")
	}
	if len([]rune(strings.TrimSpace(input.Location))) > 128 {
		return schedulePayload{}, errors.New("--location must be at most 128 characters")
	}
	adminIDs, err := parseUserIDs(input.Admins)
	if err != nil {
		return schedulePayload{}, err
	}
	if len(adminIDs) > 3 {
		return schedulePayload{}, errors.New("schedule admins can include at most 3 users")
	}
	attendees, err := buildScheduleAttendees(input.Attendees)
	if err != nil {
		return schedulePayload{}, err
	}
	reminders, err := buildScheduleReminder(input)
	if err != nil {
		return schedulePayload{}, err
	}
	payload := schedulePayload{
		ScheduleID:  strings.TrimSpace(input.ScheduleID),
		Admins:      uniqueStrings(adminIDs),
		StartTime:   startTime,
		EndTime:     endTime,
		Attendees:   attendees,
		Summary:     strings.TrimSpace(input.Summary),
		Description: strings.TrimSpace(input.Description),
		Reminders:   reminders,
		Location:    strings.TrimSpace(input.Location),
	}
	if input.WholeDay {
		payload.IsWholeDay = 1
	}
	if !update {
		payload.CalID = strings.TrimSpace(input.CalID)
	}
	return payload, nil
}

func buildScheduleReminder(input scheduleInput) (*scheduleReminder, error) {
	hasReminder := input.Remind || input.RemindBeforeEventSecs >= 0 || len(input.RemindTimeDiffs) > 0 ||
		input.Repeat || input.RepeatType >= 0 || strings.TrimSpace(input.RepeatUntil) != "" ||
		input.CustomRepeat || input.RepeatInterval > 0 || len(input.RepeatDayOfWeek) > 0 ||
		len(input.RepeatDayOfMonth) > 0 || (input.Timezone != nil && *input.Timezone != 999)
	if !hasReminder {
		return nil, nil
	}
	reminder := &scheduleReminder{}
	if input.Remind || input.RemindBeforeEventSecs >= 0 || len(input.RemindTimeDiffs) > 0 {
		reminder.IsRemind = 1
	}
	if input.RemindBeforeEventSecs >= 0 {
		reminder.RemindBeforeEventSecs = intPtr(input.RemindBeforeEventSecs)
	}
	reminder.RemindTimeDiffs = input.RemindTimeDiffs
	if input.Repeat || input.RepeatType >= 0 || strings.TrimSpace(input.RepeatUntil) != "" {
		reminder.IsRepeat = 1
	}
	if input.RepeatType >= 0 {
		if !validRepeatType(input.RepeatType) {
			return nil, errors.New("--repeat-type must be 0, 1, 2, 5, or 7")
		}
		reminder.RepeatType = intPtr(input.RepeatType)
	}
	repeatUntil, err := parseOptionalTime(input.RepeatUntil, "--repeat-until")
	if err != nil {
		return nil, err
	}
	reminder.RepeatUntil = repeatUntil
	if input.CustomRepeat || input.RepeatInterval > 0 || len(input.RepeatDayOfWeek) > 0 || len(input.RepeatDayOfMonth) > 0 {
		reminder.IsCustomRepeat = 1
	}
	if input.RepeatInterval > 0 {
		reminder.RepeatInterval = input.RepeatInterval
	}
	reminder.RepeatDayOfWeek = input.RepeatDayOfWeek
	reminder.RepeatDayOfMonth = input.RepeatDayOfMonth
	if input.Timezone != nil && *input.Timezone != 999 {
		if *input.Timezone < -12 || *input.Timezone > 12 {
			return nil, errors.New("--timezone must be between -12 and 12")
		}
		reminder.Timezone = input.Timezone
	}
	return reminder, nil
}

func buildScheduleGetRequest(scheduleIDs []string) (scheduleGetRequest, error) {
	ids := uniqueStrings(scheduleIDs)
	if len(ids) == 0 {
		return scheduleGetRequest{}, errors.New("--schedule-id is required")
	}
	if len(ids) > 1000 {
		return scheduleGetRequest{}, errors.New("schedule get can include at most 1000 schedule IDs")
	}
	return scheduleGetRequest{ScheduleIDList: ids}, nil
}

func buildScheduleListRequest(calID string, offset int, limit int) (scheduleListRequest, error) {
	calID = strings.TrimSpace(calID)
	if calID == "" {
		return scheduleListRequest{}, errors.New("--cal-id is required")
	}
	if offset < 0 {
		return scheduleListRequest{}, errors.New("--offset must be greater than or equal to 0")
	}
	if limit <= 0 || limit > 1000 {
		return scheduleListRequest{}, errors.New("--limit must be between 1 and 1000")
	}
	return scheduleListRequest{CalID: calID, Offset: offset, Limit: limit}, nil
}

func buildScheduleDeleteRequest(scheduleID string, opMode int, opStartTime string) (scheduleDeleteRequest, error) {
	scheduleID = strings.TrimSpace(scheduleID)
	if scheduleID == "" {
		return scheduleDeleteRequest{}, errors.New("--schedule-id is required")
	}
	if opMode != 0 && opMode != 1 && opMode != 2 {
		return scheduleDeleteRequest{}, errors.New("--op-mode must be 0, 1, or 2")
	}
	req := scheduleDeleteRequest{ScheduleID: scheduleID, OpMode: opMode}
	if opMode != 0 {
		parsed, err := parseOptionalTime(opStartTime, "--op-start-time")
		if err != nil {
			return scheduleDeleteRequest{}, err
		}
		if parsed == 0 {
			return scheduleDeleteRequest{}, errors.New("--op-start-time is required when --op-mode is 1 or 2")
		}
		req.OpStartTime = parsed
	}
	return req, nil
}

func buildScheduleAttendeesRequest(scheduleID string, userIDs []string) (scheduleAttendeesRequest, error) {
	scheduleID = strings.TrimSpace(scheduleID)
	if scheduleID == "" {
		return scheduleAttendeesRequest{}, errors.New("--schedule-id is required")
	}
	attendees, err := buildScheduleAttendees(userIDs)
	if err != nil {
		return scheduleAttendeesRequest{}, err
	}
	if len(attendees) == 0 {
		return scheduleAttendeesRequest{}, errors.New("--attendee is required")
	}
	return scheduleAttendeesRequest{ScheduleID: scheduleID, Attendees: attendees}, nil
}

func buildScheduleAttendees(userIDs []string) ([]scheduleAttendee, error) {
	resolved, err := parseUserIDs(userIDs)
	if err != nil {
		return nil, err
	}
	if len(resolved) > 1000 {
		return nil, errors.New("schedule attendees can include at most 1000 users")
	}
	attendees := make([]scheduleAttendee, 0, len(resolved))
	for _, userID := range resolved {
		attendees = append(attendees, scheduleAttendee{UserID: userID})
	}
	return attendees, nil
}

func parseRequiredTime(value string, flagName string) (int64, error) {
	parsed, err := parseOptionalTime(value, flagName)
	if err != nil {
		return 0, err
	}
	if parsed == 0 {
		return 0, fmt.Errorf("%s is required", flagName)
	}
	return parsed, nil
}

func parseOptionalTime(value string, flagName string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
		if parsed < 0 {
			return 0, fmt.Errorf("%s must be a positive Unix timestamp", flagName)
		}
		return parsed, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return 0, fmt.Errorf("%s must be Unix seconds or RFC3339 time", flagName)
	}
	return parsed.Unix(), nil
}

func validRepeatType(value int) bool {
	switch value {
	case 0, 1, 2, 5, 7:
		return true
	default:
		return false
	}
}

func (c *wecomClient) addSchedule(req scheduleRequest) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	path := "/cgi-bin/oa/schedule/add?access_token=" + url.QueryEscape(token)
	return c.postWeComAndTrack(path, req, resourceTrackSpec{
		Type:     "schedule",
		IDFields: []string{"schedule_id"},
		Name:     req.Schedule.Summary,
		Command:  "schedule create",
		Request:  req,
	})
}

func (c *wecomClient) updateSchedule(req scheduleUpdateRequest) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	path := "/cgi-bin/oa/schedule/update?access_token=" + url.QueryEscape(token)
	return c.postWeCom(path, req)
}

func (c *wecomClient) getSchedule(req scheduleGetRequest) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	path := "/cgi-bin/oa/schedule/get?access_token=" + url.QueryEscape(token)
	return c.postWeCom(path, req)
}

func (c *wecomClient) listSchedule(req scheduleListRequest) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	path := "/cgi-bin/oa/schedule/get_by_calendar?access_token=" + url.QueryEscape(token)
	return c.postWeCom(path, req)
}

func (c *wecomClient) deleteSchedule(req scheduleDeleteRequest) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	path := "/cgi-bin/oa/schedule/del?access_token=" + url.QueryEscape(token)
	return c.postWeCom(path, req)
}

func (c *wecomClient) addScheduleAttendees(req scheduleAttendeesRequest) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	path := "/cgi-bin/oa/schedule/add_attendees?access_token=" + url.QueryEscape(token)
	return c.postWeCom(path, req)
}

func (c *wecomClient) removeScheduleAttendees(req scheduleAttendeesRequest) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	path := "/cgi-bin/oa/schedule/del_attendees?access_token=" + url.QueryEscape(token)
	return c.postWeCom(path, req)
}

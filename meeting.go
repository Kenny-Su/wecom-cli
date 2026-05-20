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

type meetingPayload struct {
	MeetingID       string            `json:"meetingid,omitempty"`
	AdminUserID     string            `json:"admin_userid,omitempty"`
	Title           string            `json:"title,omitempty"`
	MeetingStart    int64             `json:"meeting_start,omitempty"`
	MeetingDuration int               `json:"meeting_duration,omitempty"`
	Description     string            `json:"description,omitempty"`
	Location        string            `json:"location,omitempty"`
	AgentID         int64             `json:"agentid,omitempty"`
	Invitees        *meetingUserList  `json:"invitees,omitempty"`
	Settings        *meetingSettings  `json:"settings,omitempty"`
	CalID           string            `json:"cal_id,omitempty"`
	Reminders       *meetingReminders `json:"reminders,omitempty"`
}

type meetingIDRequest struct {
	MeetingID string `json:"meetingid"`
}

type meetingUserMeetingIDRequest struct {
	UserID    string `json:"userid"`
	Cursor    string `json:"cursor,omitempty"`
	BeginTime int64  `json:"begin_time,omitempty"`
	EndTime   int64  `json:"end_time,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type meetingUserList struct {
	UserID []string `json:"userid,omitempty"`
}

type meetingSettings struct {
	RemindScope           int              `json:"remind_scope,omitempty"`
	Password              string           `json:"password,omitempty"`
	EnableWaitingRoom     *bool            `json:"enable_waiting_room,omitempty"`
	AllowEnterBeforeHost  *bool            `json:"allow_enter_before_host,omitempty"`
	EnableEnterMute       *int             `json:"enable_enter_mute,omitempty"`
	EnableScreenWatermark *bool            `json:"enable_screen_watermark,omitempty"`
	Hosts                 *meetingUserList `json:"hosts,omitempty"`
	RingUsers             *meetingUserList `json:"ring_users,omitempty"`
}

type meetingReminders struct {
	IsRepeat       int   `json:"is_repeat,omitempty"`
	RepeatType     *int  `json:"repeat_type,omitempty"`
	RepeatUntil    int64 `json:"repeat_until,omitempty"`
	RepeatInterval int   `json:"repeat_interval,omitempty"`
	RemindBefore   []int `json:"remind_before,omitempty"`
}

type meetingInput struct {
	MeetingID             string
	AdminUserID           string
	AdminName             string
	Title                 string
	MeetingStart          string
	MeetingDuration       int
	Description           string
	Location              string
	CalID                 string
	Invitees              []string
	InviteeNames          []string
	RemindScope           int
	Password              string
	EnableWaitingRoom     string
	AllowEnterBeforeHost  string
	EnableEnterMute       int
	EnableScreenWatermark string
	Hosts                 []string
	HostNames             []string
	RingUsers             []string
	RingUserNames         []string
	Repeat                bool
	RepeatType            int
	RepeatUntil           string
	RepeatInterval        int
	RemindBefore          []int
}

func runMeeting(c *wecomClient, args []string) error {
	switch args[0] {
	case "create":
		return meetingCreate(c, args[1:])
	case "update":
		return meetingUpdate(c, args[1:])
	case "get":
		return meetingGet(c, args[1:])
	case "list":
		return meetingList(c, args[1:])
	case "cancel":
		return meetingCancel(c, args[1:])
	case "help", "-h", "--help":
		printMeetingUsage()
		return nil
	default:
		return fmt.Errorf("unknown meeting command %q", args[0])
	}
}

func meetingCreate(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("meeting create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	input, dryRun := bindMeetingFlags(fs, false)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildMeetingCreateRequest(c.cfg, *input)
	if err != nil {
		return err
	}
	if *dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	return c.createMeeting(req)
}

func meetingUpdate(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("meeting update", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	input, dryRun := bindMeetingFlags(fs, true)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildMeetingUpdateRequest(c.cfg, *input)
	if err != nil {
		return err
	}
	if *dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	return c.updateMeeting(req)
}

func meetingGet(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("meeting get", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	meetingID := fs.String("meeting-id", "", "meeting ID")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildMeetingIDRequest(*meetingID)
	if err != nil {
		return err
	}
	if *dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	return c.getMeeting(req)
}

func meetingCancel(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("meeting cancel", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	meetingID := fs.String("meeting-id", "", "meeting ID")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildMeetingIDRequest(*meetingID)
	if err != nil {
		return err
	}
	if *dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	return c.cancelMeeting(req)
}

func meetingList(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("meeting list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	userID := fs.String("userid", "", "member userid")
	userName := fs.String("user-name", "", "employee name to resolve through agw-cli")
	cursor := fs.String("cursor", "", "pagination cursor")
	begin := fs.String("begin", "", "begin time, Unix seconds or RFC3339")
	end := fs.String("end", "", "end time, Unix seconds or RFC3339")
	limit := fs.Int("limit", 100, "page size, 1-100")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildMeetingListRequest(c.cfg, *userID, *userName, *cursor, *begin, *end, *limit)
	if err != nil {
		return err
	}
	if *dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	return c.listMeetings(req)
}

func bindMeetingFlags(fs *flag.FlagSet, includeID bool) (*meetingInput, *bool) {
	input := &meetingInput{}
	if includeID {
		fs.StringVar(&input.MeetingID, "meeting-id", "", "meeting ID")
	}
	fs.StringVar(&input.AdminUserID, "admin-userid", "", "meeting admin userid")
	fs.StringVar(&input.AdminName, "admin-name", "", "meeting admin name to resolve through agw-cli")
	fs.StringVar(&input.Title, "title", "", "meeting title")
	fs.StringVar(&input.MeetingStart, "start", "", "meeting start time, Unix seconds or RFC3339")
	fs.IntVar(&input.MeetingDuration, "duration", 0, "meeting duration in seconds")
	fs.StringVar(&input.Description, "description", "", "meeting description")
	fs.StringVar(&input.Location, "location", "", "meeting location")
	fs.StringVar(&input.CalID, "cal-id", "", "calendar ID")
	fs.Var((*stringList)(&input.Invitees), "invitee", "invitee userid; repeatable")
	fs.Var((*stringList)(&input.InviteeNames), "invitee-name", "invitee employee name; repeatable")
	fs.IntVar(&input.RemindScope, "remind-scope", 0, "1 none, 2 hosts, 3 all, 4 ring-users")
	fs.StringVar(&input.Password, "password", "", "meeting password, 4-6 digits")
	fs.StringVar(&input.EnableWaitingRoom, "waiting-room", "", "true or false")
	fs.StringVar(&input.AllowEnterBeforeHost, "allow-enter-before-host", "", "true or false")
	fs.IntVar(&input.EnableEnterMute, "enter-mute", -1, "0 off, 1 on, 2 auto")
	fs.StringVar(&input.EnableScreenWatermark, "screen-watermark", "", "true or false")
	fs.Var((*stringList)(&input.Hosts), "host", "host userid; repeatable")
	fs.Var((*stringList)(&input.HostNames), "host-name", "host employee name; repeatable")
	fs.Var((*stringList)(&input.RingUsers), "ring-user", "ring userid; repeatable")
	fs.Var((*stringList)(&input.RingUserNames), "ring-user-name", "ring employee name; repeatable")
	fs.BoolVar(&input.Repeat, "repeat", false, "enable repeat meeting")
	fs.IntVar(&input.RepeatType, "repeat-type", -1, "0 daily, 1 weekly, 2 monthly, 7 workday")
	fs.StringVar(&input.RepeatUntil, "repeat-until", "", "repeat end time, Unix seconds or RFC3339")
	fs.IntVar(&input.RepeatInterval, "repeat-interval", 0, "repeat interval")
	fs.Var((*intList)(&input.RemindBefore), "remind-before", "meeting reminder before start in seconds; repeatable")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	return input, dryRun
}

func buildMeetingCreateRequest(cfg config, input meetingInput) (meetingPayload, error) {
	if strings.TrimSpace(input.AdminUserID) == "" && strings.TrimSpace(input.AdminName) == "" {
		return meetingPayload{}, errors.New("--admin-userid or --admin-name is required")
	}
	if strings.TrimSpace(input.Title) == "" {
		return meetingPayload{}, errors.New("--title is required")
	}
	if input.MeetingDuration == 0 {
		return meetingPayload{}, errors.New("--duration is required")
	}
	payload, err := buildMeetingPayload(cfg, input, false)
	if err != nil {
		return meetingPayload{}, err
	}
	payload.AgentID = cfg.AgentID
	return payload, nil
}

func buildMeetingUpdateRequest(cfg config, input meetingInput) (meetingPayload, error) {
	if strings.TrimSpace(input.MeetingID) == "" {
		return meetingPayload{}, errors.New("--meeting-id is required")
	}
	if (strings.TrimSpace(input.MeetingStart) == "") != (input.MeetingDuration == 0) {
		return meetingPayload{}, errors.New("--start and --duration must be specified together")
	}
	return buildMeetingPayload(cfg, input, true)
}

func buildMeetingPayload(cfg config, input meetingInput, update bool) (meetingPayload, error) {
	adminUserID := strings.TrimSpace(input.AdminUserID)
	if adminUserID == "" && strings.TrimSpace(input.AdminName) != "" {
		resolved, err := resolveUserName(cfg, input.AdminName)
		if err != nil {
			return meetingPayload{}, err
		}
		adminUserID = resolved
	}
	title := strings.TrimSpace(input.Title)
	if title != "" && (len([]byte(title)) > 40 || len([]rune(title)) > 20) {
		return meetingPayload{}, errors.New("--title must be at most 40 bytes or 20 UTF-8 characters")
	}
	description := strings.TrimSpace(input.Description)
	if len([]rune(description)) > 500 {
		return meetingPayload{}, errors.New("--description must be at most 500 characters")
	}
	location := strings.TrimSpace(input.Location)
	if len([]rune(location)) > 128 {
		return meetingPayload{}, errors.New("--location must be at most 128 characters")
	}
	start, err := parseOptionalTime(input.MeetingStart, "--start")
	if err != nil {
		return meetingPayload{}, err
	}
	if !update && start == 0 {
		return meetingPayload{}, errors.New("--start is required")
	}
	if input.MeetingDuration != 0 && (input.MeetingDuration < 300 || input.MeetingDuration > 86399) {
		return meetingPayload{}, errors.New("--duration must be between 300 and 86399 seconds")
	}
	invitees, err := buildMeetingUserList(cfg, input.Invitees, input.InviteeNames)
	if err != nil {
		return meetingPayload{}, err
	}
	settings, err := buildMeetingSettings(cfg, input)
	if err != nil {
		return meetingPayload{}, err
	}
	reminders, err := buildMeetingReminders(input)
	if err != nil {
		return meetingPayload{}, err
	}
	payload := meetingPayload{
		MeetingID:       strings.TrimSpace(input.MeetingID),
		AdminUserID:     adminUserID,
		Title:           title,
		MeetingStart:    start,
		MeetingDuration: input.MeetingDuration,
		Description:     description,
		Location:        location,
		Invitees:        invitees,
		Settings:        settings,
		CalID:           strings.TrimSpace(input.CalID),
		Reminders:       reminders,
	}
	return payload, nil
}

func buildMeetingSettings(cfg config, input meetingInput) (*meetingSettings, error) {
	hosts, err := buildMeetingUserList(cfg, input.Hosts, input.HostNames)
	if err != nil {
		return nil, err
	}
	ringUsers, err := buildMeetingUserList(cfg, input.RingUsers, input.RingUserNames)
	if err != nil {
		return nil, err
	}
	waitingRoom, err := parseOptionalBool(input.EnableWaitingRoom, "--waiting-room")
	if err != nil {
		return nil, err
	}
	enterBeforeHost, err := parseOptionalBool(input.AllowEnterBeforeHost, "--allow-enter-before-host")
	if err != nil {
		return nil, err
	}
	watermark, err := parseOptionalBool(input.EnableScreenWatermark, "--screen-watermark")
	if err != nil {
		return nil, err
	}
	if input.RemindScope < 0 || input.RemindScope > 4 {
		return nil, errors.New("--remind-scope must be between 1 and 4")
	}
	if input.EnableEnterMute != -1 && (input.EnableEnterMute < 0 || input.EnableEnterMute > 2) {
		return nil, errors.New("--enter-mute must be 0, 1, or 2")
	}
	if strings.TrimSpace(input.Password) != "" && !validMeetingPassword(input.Password) {
		return nil, errors.New("--password must be 4 to 6 digits")
	}
	hasSettings := input.RemindScope > 0 || strings.TrimSpace(input.Password) != "" ||
		waitingRoom != nil || enterBeforeHost != nil || input.EnableEnterMute != -1 ||
		watermark != nil || hosts != nil || ringUsers != nil
	if !hasSettings {
		return nil, nil
	}
	settings := &meetingSettings{
		RemindScope:           input.RemindScope,
		Password:              strings.TrimSpace(input.Password),
		EnableWaitingRoom:     waitingRoom,
		AllowEnterBeforeHost:  enterBeforeHost,
		EnableScreenWatermark: watermark,
		Hosts:                 hosts,
		RingUsers:             ringUsers,
	}
	if input.EnableEnterMute != -1 {
		settings.EnableEnterMute = intPtr(input.EnableEnterMute)
	}
	return settings, nil
}

func buildMeetingReminders(input meetingInput) (*meetingReminders, error) {
	hasReminders := input.Repeat || input.RepeatType >= 0 || strings.TrimSpace(input.RepeatUntil) != "" ||
		input.RepeatInterval > 0 || len(input.RemindBefore) > 0
	if !hasReminders {
		return nil, nil
	}
	reminders := &meetingReminders{}
	if input.Repeat || input.RepeatType >= 0 || strings.TrimSpace(input.RepeatUntil) != "" {
		reminders.IsRepeat = 1
	}
	if input.RepeatType >= 0 {
		if !validMeetingRepeatType(input.RepeatType) {
			return nil, errors.New("--repeat-type must be 0, 1, 2, or 7")
		}
		reminders.RepeatType = intPtr(input.RepeatType)
	}
	repeatUntil, err := parseOptionalTime(input.RepeatUntil, "--repeat-until")
	if err != nil {
		return nil, err
	}
	reminders.RepeatUntil = repeatUntil
	if input.RepeatInterval > 0 {
		reminders.RepeatInterval = input.RepeatInterval
	}
	reminders.RemindBefore = input.RemindBefore
	return reminders, nil
}

func buildMeetingUserList(cfg config, userIDs []string, names []string) (*meetingUserList, error) {
	resolved, err := resolveUsers(cfg, userIDs, names)
	if err != nil {
		return nil, err
	}
	resolved = uniqueStrings(resolved)
	if len(resolved) == 0 {
		return nil, nil
	}
	if len(resolved) > 300 {
		return nil, errors.New("meeting user list can include at most 300 users")
	}
	return &meetingUserList{UserID: resolved}, nil
}

func buildMeetingIDRequest(meetingID string) (meetingIDRequest, error) {
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return meetingIDRequest{}, errors.New("--meeting-id is required")
	}
	return meetingIDRequest{MeetingID: meetingID}, nil
}

func buildMeetingListRequest(cfg config, userID string, userName string, cursor string, begin string, end string, limit int) (meetingUserMeetingIDRequest, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" && strings.TrimSpace(userName) != "" {
		resolved, err := resolveUserName(cfg, userName)
		if err != nil {
			return meetingUserMeetingIDRequest{}, err
		}
		userID = resolved
	}
	if userID == "" {
		return meetingUserMeetingIDRequest{}, errors.New("--userid or --user-name is required")
	}
	if limit <= 0 || limit > 100 {
		return meetingUserMeetingIDRequest{}, errors.New("--limit must be between 1 and 100")
	}
	beginTime, err := parseOptionalTime(begin, "--begin")
	if err != nil {
		return meetingUserMeetingIDRequest{}, err
	}
	endTime, err := parseOptionalTime(end, "--end")
	if err != nil {
		return meetingUserMeetingIDRequest{}, err
	}
	if beginTime != 0 && endTime != 0 && endTime <= beginTime {
		return meetingUserMeetingIDRequest{}, errors.New("--end must be after --begin")
	}
	return meetingUserMeetingIDRequest{
		UserID:    userID,
		Cursor:    strings.TrimSpace(cursor),
		BeginTime: beginTime,
		EndTime:   endTime,
		Limit:     limit,
	}, nil
}

func parseOptionalBool(value string, flagName string) (*bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, fmt.Errorf("%s must be true or false", flagName)
	}
	return boolPtr(parsed), nil
}

func validMeetingPassword(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) < 4 || len(value) > 6 {
		return false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func validMeetingRepeatType(value int) bool {
	switch value {
	case 0, 1, 2, 7:
		return true
	default:
		return false
	}
}

func (c *wecomClient) createMeeting(req meetingPayload) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	path := "/cgi-bin/meeting/create?access_token=" + url.QueryEscape(token)
	return c.postWeCom(path, req)
}

func (c *wecomClient) updateMeeting(req meetingPayload) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	path := "/cgi-bin/meeting/update?access_token=" + url.QueryEscape(token)
	return c.postWeCom(path, req)
}

func (c *wecomClient) getMeeting(req meetingIDRequest) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	path := "/cgi-bin/meeting/get_info?access_token=" + url.QueryEscape(token)
	return c.postWeCom(path, req)
}

func (c *wecomClient) listMeetings(req meetingUserMeetingIDRequest) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	path := "/cgi-bin/meeting/get_user_meetingid?access_token=" + url.QueryEscape(token)
	return c.postWeCom(path, req)
}

func (c *wecomClient) cancelMeeting(req meetingIDRequest) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	path := "/cgi-bin/meeting/cancel?access_token=" + url.QueryEscape(token)
	return c.postWeCom(path, req)
}

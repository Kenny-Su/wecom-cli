package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"net/url"
	"strconv"
	"strings"
)

type weDriveSpaceCreateRequest struct {
	SpaceName    string            `json:"space_name"`
	AuthInfo     []weDriveAuthInfo `json:"auth_info,omitempty"`
	SpaceSubType int               `json:"space_sub_type,omitempty"`
}

type weDriveSpaceIDRequest struct {
	SpaceID string `json:"spaceid"`
}

type weDriveSpaceRenameRequest struct {
	SpaceID   string `json:"spaceid"`
	SpaceName string `json:"space_name"`
}

type weDriveSpaceACLRequest struct {
	SpaceID  string            `json:"spaceid"`
	AuthInfo []weDriveAuthInfo `json:"auth_info"`
}

type weDriveAuthInfo struct {
	Type         int    `json:"type"`
	UserID       string `json:"userid,omitempty"`
	DepartmentID uint32 `json:"departmentid,omitempty"`
	Auth         int    `json:"auth,omitempty"`
}

type weDriveSpaceSettingRequest struct {
	SpaceID                      string `json:"spaceid"`
	EnableWatermark              *bool  `json:"enable_watermark,omitempty"`
	ShareURLNoApprove            *bool  `json:"share_url_no_approve,omitempty"`
	ShareURLNoApproveDefaultAuth int    `json:"share_url_no_approve_default_auth,omitempty"`
	EnableConfidentialMode       *bool  `json:"enable_confidential_mode,omitempty"`
	DefaultFileScope             int    `json:"default_file_scope,omitempty"`
	BanShareExternal             *bool  `json:"ban_share_external,omitempty"`
}

func runWeDrive(c *wecomClient, args []string) error {
	switch args[0] {
	case "space":
		if len(args) == 1 || isHelp(args[1]) {
			printWeDriveSpaceUsage()
			return nil
		}
		return runWeDriveSpace(c, args[1:])
	case "file":
		if len(args) == 1 || isHelp(args[1]) {
			printWeDriveFileUsage()
			return nil
		}
		return runWeDriveFile(c, args[1:])
	case "help", "-h", "--help":
		printWeDriveUsage()
		return nil
	default:
		return fmt.Errorf("unknown wedrive command %q", args[0])
	}
}

func runWeDriveSpace(c *wecomClient, args []string) error {
	switch args[0] {
	case "create":
		return weDriveSpaceCreate(c, args[1:])
	case "info":
		return weDriveSpaceInfo(c, args[1:], false)
	case "new-info", "permission-info":
		return weDriveSpaceInfo(c, args[1:], true)
	case "rename":
		return weDriveSpaceRename(c, args[1:])
	case "dismiss":
		return weDriveSpaceDismiss(c, args[1:])
	case "share":
		return weDriveSpaceShare(c, args[1:])
	case "acl-add":
		return weDriveSpaceACL(c, args[1:], true)
	case "acl-del", "acl-remove":
		return weDriveSpaceACL(c, args[1:], false)
	case "setting":
		return weDriveSpaceSetting(c, args[1:])
	case "help", "-h", "--help":
		printWeDriveSpaceUsage()
		return nil
	default:
		return fmt.Errorf("unknown wedrive space command %q", args[0])
	}
}

func weDriveSpaceCreate(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("wedrive space create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	spaceName := fs.String("space-name", "", "space name")
	spaceSubType := fs.Int("space-sub-type", 0, "space subtype; currently only 0")
	var members, departments stringList
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	fs.Var(&members, "member", "userid:auth; repeatable")
	fs.Var(&departments, "department", "departmentid:auth; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	authInfo, err := buildWeDriveAuthInfo(members, departments, true)
	if err != nil {
		return err
	}
	req, err := buildWeDriveSpaceCreateRequest(*spaceName, *spaceSubType, authInfo)
	if err != nil {
		return err
	}
	return c.runWeDriveRequest(*dryRun, req, c.createWeDriveSpace)
}

func weDriveSpaceInfo(c *wecomClient, args []string, withPermissions bool) error {
	fs := flag.NewFlagSet("wedrive space info", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	spaceID := fs.String("spaceid", "", "space ID")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildWeDriveSpaceIDRequest(*spaceID)
	if err != nil {
		return err
	}
	if withPermissions {
		return c.runWeDriveRequest(*dryRun, req, c.getWeDriveNewSpaceInfo)
	}
	return c.runWeDriveRequest(*dryRun, req, c.getWeDriveSpaceInfo)
}

func weDriveSpaceRename(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("wedrive space rename", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	spaceID := fs.String("spaceid", "", "space ID")
	spaceName := fs.String("space-name", "", "new space name")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildWeDriveSpaceRenameRequest(*spaceID, *spaceName)
	if err != nil {
		return err
	}
	return c.runWeDriveRequest(*dryRun, req, c.renameWeDriveSpace)
}

func weDriveSpaceDismiss(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("wedrive space dismiss", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	spaceID := fs.String("spaceid", "", "space ID")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildWeDriveSpaceIDRequest(*spaceID)
	if err != nil {
		return err
	}
	return c.runWeDriveRequest(*dryRun, req, c.dismissWeDriveSpace)
}

func weDriveSpaceShare(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("wedrive space share", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	spaceID := fs.String("spaceid", "", "space ID")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildWeDriveSpaceIDRequest(*spaceID)
	if err != nil {
		return err
	}
	return c.runWeDriveRequest(*dryRun, req, c.shareWeDriveSpace)
}

func weDriveSpaceACL(c *wecomClient, args []string, add bool) error {
	name := "wedrive space acl-del"
	if add {
		name = "wedrive space acl-add"
	}
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	spaceID := fs.String("spaceid", "", "space ID")
	var members, departments stringList
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	fs.Var(&members, "member", "userid[:auth]; repeatable")
	fs.Var(&departments, "department", "departmentid[:auth]; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	authInfo, err := buildWeDriveAuthInfo(members, departments, add)
	if err != nil {
		return err
	}
	req, err := buildWeDriveSpaceACLRequest(*spaceID, authInfo)
	if err != nil {
		return err
	}
	if add {
		return c.runWeDriveRequest(*dryRun, req, c.addWeDriveSpaceACL)
	}
	return c.runWeDriveRequest(*dryRun, req, c.removeWeDriveSpaceACL)
}

func weDriveSpaceSetting(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("wedrive space setting", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	spaceID := fs.String("spaceid", "", "space ID")
	enableWatermark := fs.String("enable-watermark", "", "true or false")
	shareURLNoApprove := fs.String("share-url-no-approve", "", "true or false")
	shareURLNoApproveAuth := fs.Int("share-url-no-approve-default-auth", 0, "invite link default auth")
	enableConfidentialMode := fs.String("enable-confidential-mode", "", "true or false")
	defaultFileScope := fs.Int("default-file-scope", 0, "1 members only, 2 corp")
	banShareExternal := fs.String("ban-share-external", "", "true or false")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildWeDriveSpaceSettingRequest(*spaceID, *enableWatermark, *shareURLNoApprove, *shareURLNoApproveAuth, *enableConfidentialMode, *defaultFileScope, *banShareExternal)
	if err != nil {
		return err
	}
	return c.runWeDriveRequest(*dryRun, req, c.setWeDriveSpaceSetting)
}

func buildWeDriveSpaceCreateRequest(spaceName string, spaceSubType int, authInfo []weDriveAuthInfo) (weDriveSpaceCreateRequest, error) {
	spaceName = strings.TrimSpace(spaceName)
	if spaceName == "" {
		return weDriveSpaceCreateRequest{}, errors.New("--space-name is required")
	}
	if spaceSubType != 0 {
		return weDriveSpaceCreateRequest{}, errors.New("--space-sub-type must be 0")
	}
	return weDriveSpaceCreateRequest{SpaceName: spaceName, AuthInfo: authInfo, SpaceSubType: spaceSubType}, nil
}

func buildWeDriveSpaceIDRequest(spaceID string) (weDriveSpaceIDRequest, error) {
	spaceID = strings.TrimSpace(spaceID)
	if spaceID == "" {
		return weDriveSpaceIDRequest{}, errors.New("--spaceid is required")
	}
	return weDriveSpaceIDRequest{SpaceID: spaceID}, nil
}

func buildWeDriveSpaceRenameRequest(spaceID string, spaceName string) (weDriveSpaceRenameRequest, error) {
	idReq, err := buildWeDriveSpaceIDRequest(spaceID)
	if err != nil {
		return weDriveSpaceRenameRequest{}, err
	}
	spaceName = strings.TrimSpace(spaceName)
	if spaceName == "" {
		return weDriveSpaceRenameRequest{}, errors.New("--space-name is required")
	}
	return weDriveSpaceRenameRequest{SpaceID: idReq.SpaceID, SpaceName: spaceName}, nil
}

func buildWeDriveSpaceACLRequest(spaceID string, authInfo []weDriveAuthInfo) (weDriveSpaceACLRequest, error) {
	idReq, err := buildWeDriveSpaceIDRequest(spaceID)
	if err != nil {
		return weDriveSpaceACLRequest{}, err
	}
	if len(authInfo) == 0 {
		return weDriveSpaceACLRequest{}, errors.New("--member or --department is required")
	}
	return weDriveSpaceACLRequest{SpaceID: idReq.SpaceID, AuthInfo: authInfo}, nil
}

func buildWeDriveSpaceSettingRequest(spaceID string, enableWatermark string, shareURLNoApprove string, shareURLNoApproveAuth int, enableConfidentialMode string, defaultFileScope int, banShareExternal string) (weDriveSpaceSettingRequest, error) {
	idReq, err := buildWeDriveSpaceIDRequest(spaceID)
	if err != nil {
		return weDriveSpaceSettingRequest{}, err
	}
	watermark, err := parseOptionalBool(enableWatermark, "--enable-watermark")
	if err != nil {
		return weDriveSpaceSettingRequest{}, err
	}
	noApprove, err := parseOptionalBool(shareURLNoApprove, "--share-url-no-approve")
	if err != nil {
		return weDriveSpaceSettingRequest{}, err
	}
	confidential, err := parseOptionalBool(enableConfidentialMode, "--enable-confidential-mode")
	if err != nil {
		return weDriveSpaceSettingRequest{}, err
	}
	banExternal, err := parseOptionalBool(banShareExternal, "--ban-share-external")
	if err != nil {
		return weDriveSpaceSettingRequest{}, err
	}
	if shareURLNoApproveAuth != 0 && !validWeDriveShareDefaultAuth(shareURLNoApproveAuth) {
		return weDriveSpaceSettingRequest{}, errors.New("--share-url-no-approve-default-auth must be 1, 2, 4, 5, or 200")
	}
	if defaultFileScope != 0 && defaultFileScope != 1 && defaultFileScope != 2 {
		return weDriveSpaceSettingRequest{}, errors.New("--default-file-scope must be 1 or 2")
	}
	return weDriveSpaceSettingRequest{
		SpaceID:                      idReq.SpaceID,
		EnableWatermark:              watermark,
		ShareURLNoApprove:            noApprove,
		ShareURLNoApproveDefaultAuth: shareURLNoApproveAuth,
		EnableConfidentialMode:       confidential,
		DefaultFileScope:             defaultFileScope,
		BanShareExternal:             banExternal,
	}, nil
}

func buildWeDriveAuthInfo(members []string, departments []string, requireAuth bool) ([]weDriveAuthInfo, error) {
	var out []weDriveAuthInfo
	for _, raw := range members {
		value, auth, hasAuth, err := parseWeDriveAuthSpec(raw)
		if err != nil {
			return nil, err
		}
		if requireAuth && !hasAuth {
			return nil, fmt.Errorf("member %q requires auth", raw)
		}
		if requireAuth && !validWeDriveAuth(auth, false) {
			return nil, fmt.Errorf("member %q has invalid auth", raw)
		}
		out = append(out, weDriveAuthInfo{Type: 1, UserID: value, Auth: auth})
	}
	for _, raw := range departments {
		value, auth, hasAuth, err := parseWeDriveAuthSpec(raw)
		if err != nil {
			return nil, err
		}
		if requireAuth && !hasAuth {
			return nil, fmt.Errorf("department %q requires auth", raw)
		}
		if requireAuth && !validWeDriveAuth(auth, true) {
			return nil, fmt.Errorf("department %q has invalid auth", raw)
		}
		id, err := parseUint32(value, "departmentid")
		if err != nil {
			return nil, err
		}
		out = append(out, weDriveAuthInfo{Type: 2, DepartmentID: id, Auth: auth})
	}
	return out, nil
}

func parseWeDriveAuthSpec(raw string) (string, int, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", 0, false, errors.New("empty auth spec")
	}
	value, authText, hasAuth := strings.Cut(raw, ":")
	value = strings.TrimSpace(value)
	if value == "" {
		return "", 0, false, fmt.Errorf("invalid auth spec %q", raw)
	}
	if !hasAuth || strings.TrimSpace(authText) == "" {
		return value, 0, false, nil
	}
	auth, err := strconv.Atoi(strings.TrimSpace(authText))
	if err != nil {
		return "", 0, false, fmt.Errorf("invalid auth in %q", raw)
	}
	return value, auth, true, nil
}

func parseUint32(value string, name string) (uint32, error) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 32)
	if err != nil || parsed > math.MaxUint32 {
		return 0, fmt.Errorf("%s must be a uint32", name)
	}
	return uint32(parsed), nil
}

func validWeDriveAuth(auth int, department bool) bool {
	if department {
		return auth == 1
	}
	return auth == 1 || auth == 4 || auth == 7
}

func validWeDriveShareDefaultAuth(auth int) bool {
	switch auth {
	case 1, 2, 4, 5, 200:
		return true
	default:
		return false
	}
}

func (c *wecomClient) runWeDriveRequest(dryRun bool, req any, call func(any) error) error {
	if dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	return call(req)
}

func (c *wecomClient) postWeDrive(path string, req any) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	return c.postWeCom(path+"?access_token="+url.QueryEscape(token), req)
}

func (c *wecomClient) createWeDriveSpace(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/space_create", req)
}

func (c *wecomClient) getWeDriveSpaceInfo(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/space_info", req)
}

func (c *wecomClient) getWeDriveNewSpaceInfo(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/new_space_info", req)
}

func (c *wecomClient) renameWeDriveSpace(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/space_rename", req)
}

func (c *wecomClient) dismissWeDriveSpace(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/space_dismiss", req)
}

func (c *wecomClient) shareWeDriveSpace(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/space_share", req)
}

func (c *wecomClient) addWeDriveSpaceACL(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/space_acl_add", req)
}

func (c *wecomClient) removeWeDriveSpaceACL(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/space_acl_del", req)
}

func (c *wecomClient) setWeDriveSpaceSetting(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/space_setting", req)
}

package main

import (
	"bytes"
	"crypto/sha1"
	"encoding"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const weDriveUploadBlockSize = 2 * 1024 * 1024

type weDriveFileIDRequest struct {
	FileID string `json:"fileid"`
}

type weDriveFileListRequest struct {
	SpaceID  string `json:"spaceid"`
	FatherID string `json:"fatherid"`
	SortType int    `json:"sort_type"`
	Start    int    `json:"start"`
	Limit    int    `json:"limit"`
}

type weDriveFileCreateRequest struct {
	SpaceID  string `json:"spaceid"`
	FatherID string `json:"fatherid"`
	FileType int    `json:"file_type"`
	FileName string `json:"file_name"`
}

type weDriveFileUploadRequest struct {
	SpaceID           string `json:"spaceid,omitempty"`
	FatherID          string `json:"fatherid,omitempty"`
	SelectedTicket    string `json:"selected_ticket,omitempty"`
	FileName          string `json:"file_name"`
	FileBase64Content string `json:"file_base64_content"`
}

type weDriveFileDownloadRequest struct {
	FileID         string `json:"fileid,omitempty"`
	SelectedTicket string `json:"selected_ticket,omitempty"`
}

type weDriveFileRenameRequest struct {
	FileID  string `json:"fileid"`
	NewName string `json:"new_name"`
}

type weDriveFileMoveRequest struct {
	FatherID string   `json:"fatherid"`
	Replace  *bool    `json:"replace,omitempty"`
	FileID   []string `json:"fileid"`
}

type weDriveFileDeleteRequest struct {
	FileID []string `json:"fileid"`
}

type weDriveFileACLRequest struct {
	FileID   string            `json:"fileid"`
	AuthInfo []weDriveAuthInfo `json:"auth_info"`
}

type weDriveFileSettingRequest struct {
	FileID    string `json:"fileid"`
	AuthScope int    `json:"auth_scope"`
	Auth      int    `json:"auth,omitempty"`
}

type weDriveWatermarkSetting struct {
	Text            string `json:"text,omitempty"`
	MarginType      int    `json:"margin_type,omitempty"`
	ShowVisitorName *bool  `json:"show_visitor_name,omitempty"`
	ShowText        *bool  `json:"show_text,omitempty"`
}

type weDriveFileSecureSettingRequest struct {
	FileID    string                  `json:"fileid"`
	Watermark weDriveWatermarkSetting `json:"watermark"`
}

type weDriveFileUploadInitRequest struct {
	SpaceID        string   `json:"spaceid,omitempty"`
	FatherID       string   `json:"fatherid,omitempty"`
	SelectedTicket string   `json:"selected_ticket,omitempty"`
	FileName       string   `json:"file_name"`
	Size           uint64   `json:"size"`
	BlockSHA       []string `json:"block_sha"`
	SkipPushCard   *bool    `json:"skip_push_card,omitempty"`
}

type weDriveFileUploadPartRequest struct {
	UploadKey         string `json:"upload_key"`
	Index             int    `json:"index"`
	FileBase64Content string `json:"file_base64_content"`
}

type weDriveFileUploadFinishRequest struct {
	UploadKey string `json:"upload_key"`
}

type weDriveFileUploadInitResponse struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	HitExist  bool   `json:"hit_exist"`
	UploadKey string `json:"upload_key"`
	FileID    string `json:"fileid"`
}

func runWeDriveFile(c *wecomClient, args []string) error {
	switch args[0] {
	case "list":
		return weDriveFileList(c, args[1:])
	case "info":
		return weDriveFileInfo(c, args[1:])
	case "create":
		return weDriveFileCreate(c, args[1:])
	case "upload":
		return weDriveFileUpload(c, args[1:])
	case "upload-chunk":
		return weDriveFileUploadChunk(c, args[1:])
	case "download":
		return weDriveFileDownload(c, args[1:])
	case "rename":
		return weDriveFileRename(c, args[1:])
	case "move":
		return weDriveFileMove(c, args[1:])
	case "delete":
		return weDriveFileDelete(c, args[1:])
	case "share":
		return weDriveFileShare(c, args[1:])
	case "permission":
		return weDriveFilePermission(c, args[1:])
	case "acl-add":
		return weDriveFileACL(c, args[1:], true)
	case "acl-del", "acl-remove":
		return weDriveFileACL(c, args[1:], false)
	case "setting":
		return weDriveFileSetting(c, args[1:])
	case "secure-setting":
		return weDriveFileSecureSetting(c, args[1:])
	case "help", "-h", "--help":
		printWeDriveFileUsage()
		return nil
	default:
		return fmt.Errorf("unknown wedrive file command %q", args[0])
	}
}

func weDriveFileList(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("wedrive file list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	spaceID := fs.String("spaceid", "", "space ID")
	fatherID := fs.String("fatherid", "", "folder file ID; root uses space ID")
	sortType := fs.Int("sort-type", 1, "sort type")
	start := fs.Int("start", 0, "pagination start")
	limit := fs.Int("limit", 100, "pagination limit")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildWeDriveFileListRequest(*spaceID, *fatherID, *sortType, *start, *limit)
	if err != nil {
		return err
	}
	return c.runWeDriveRequest(*dryRun, req, c.listWeDriveFiles)
}

func weDriveFileInfo(c *wecomClient, args []string) error {
	return weDriveFileIDCommand(c, "wedrive file info", args, c.getWeDriveFileInfo)
}

func weDriveFileCreate(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("wedrive file create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	spaceID := fs.String("spaceid", "", "space ID")
	fatherID := fs.String("fatherid", "", "parent folder file ID")
	fileType := fs.Int("file-type", 0, "1 folder, 3 doc, 4 sheet")
	fileName := fs.String("file-name", "", "file name")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildWeDriveFileCreateRequest(*spaceID, *fatherID, *fileType, *fileName)
	if err != nil {
		return err
	}
	return c.runWeDriveRequest(*dryRun, req, c.createWeDriveFile)
}

func weDriveFileUpload(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("wedrive file upload", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	spaceID := fs.String("spaceid", "", "space ID")
	fatherID := fs.String("fatherid", "", "parent folder file ID")
	selectedTicket := fs.String("selected-ticket", "", "selectedTicket from JSAPI")
	fileName := fs.String("file-name", "", "file name")
	path := fs.String("path", "", "local file path to upload")
	base64Content := fs.String("base64", "", "base64 file content")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildWeDriveFileUploadRequest(*spaceID, *fatherID, *selectedTicket, *fileName, *path, *base64Content)
	if err != nil {
		return err
	}
	return c.runWeDriveRequest(*dryRun, req, c.uploadWeDriveFile)
}

func weDriveFileUploadChunk(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("wedrive file upload-chunk", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	spaceID := fs.String("spaceid", "", "space ID")
	fatherID := fs.String("fatherid", "", "parent folder file ID")
	selectedTicket := fs.String("selected-ticket", "", "selectedTicket from JSAPI")
	fileName := fs.String("file-name", "", "file name; defaults to path basename")
	path := fs.String("path", "", "local file path to upload")
	skipPushCard := fs.String("skip-push-card", "", "true or false")
	dryRun := fs.Bool("dry-run", false, "print init request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildWeDriveFileUploadInitRequest(*spaceID, *fatherID, *selectedTicket, *fileName, *path, *skipPushCard)
	if err != nil {
		return err
	}
	if *dryRun {
		return printPrettyJSON(req)
	}
	if err := c.requireCredentials(); err != nil {
		return err
	}
	return c.uploadWeDriveFileInChunks(req, *path)
}

func weDriveFileDownload(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("wedrive file download", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fileID := fs.String("fileid", "", "file ID")
	selectedTicket := fs.String("selected-ticket", "", "selectedTicket from JSAPI")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildWeDriveFileDownloadRequest(*fileID, *selectedTicket)
	if err != nil {
		return err
	}
	return c.runWeDriveRequest(*dryRun, req, c.downloadWeDriveFile)
}

func weDriveFileRename(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("wedrive file rename", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fileID := fs.String("fileid", "", "file ID")
	newName := fs.String("new-name", "", "new file name")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildWeDriveFileRenameRequest(*fileID, *newName)
	if err != nil {
		return err
	}
	return c.runWeDriveRequest(*dryRun, req, c.renameWeDriveFile)
}

func weDriveFileMove(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("wedrive file move", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fatherID := fs.String("fatherid", "", "target folder file ID")
	replace := fs.String("replace", "", "true or false")
	var fileIDs stringList
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	fs.Var(&fileIDs, "fileid", "file ID; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildWeDriveFileMoveRequest(*fatherID, *replace, fileIDs)
	if err != nil {
		return err
	}
	return c.runWeDriveRequest(*dryRun, req, c.moveWeDriveFile)
}

func weDriveFileDelete(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("wedrive file delete", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var fileIDs stringList
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	fs.Var(&fileIDs, "fileid", "file ID; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildWeDriveFileDeleteRequest(fileIDs)
	if err != nil {
		return err
	}
	return c.runWeDriveRequest(*dryRun, req, c.deleteWeDriveFile)
}

func weDriveFileShare(c *wecomClient, args []string) error {
	return weDriveFileIDCommand(c, "wedrive file share", args, c.shareWeDriveFile)
}

func weDriveFilePermission(c *wecomClient, args []string) error {
	return weDriveFileIDCommand(c, "wedrive file permission", args, c.getWeDriveFilePermission)
}

func weDriveFileACL(c *wecomClient, args []string, add bool) error {
	name := "wedrive file acl-del"
	if add {
		name = "wedrive file acl-add"
	}
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fileID := fs.String("fileid", "", "file ID")
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
	authInfo, err := buildWeDriveFileAuthInfo(members, departments, add)
	if err != nil {
		return err
	}
	req, err := buildWeDriveFileACLRequest(*fileID, authInfo)
	if err != nil {
		return err
	}
	if add {
		return c.runWeDriveRequest(*dryRun, req, c.addWeDriveFileACL)
	}
	return c.runWeDriveRequest(*dryRun, req, c.removeWeDriveFileACL)
}

func weDriveFileSetting(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("wedrive file setting", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fileID := fs.String("fileid", "", "file ID")
	authScope := fs.Int("auth-scope", 0, "1 specified users, 2 corp internal, 3 corp external, 4 internal approval, 5 external approval")
	auth := fs.Int("auth", 0, "1 download/browse, 4 preview")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildWeDriveFileSettingRequest(*fileID, *authScope, *auth)
	if err != nil {
		return err
	}
	return c.runWeDriveRequest(*dryRun, req, c.setWeDriveFileSetting)
}

func weDriveFileSecureSetting(c *wecomClient, args []string) error {
	fs := flag.NewFlagSet("wedrive file secure-setting", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fileID := fs.String("fileid", "", "file ID")
	text := fs.String("watermark-text", "", "watermark text")
	marginType := fs.Int("margin-type", 0, "1 low-density, 2 high-density")
	showVisitorName := fs.String("show-visitor-name", "", "true or false")
	showText := fs.String("show-text", "", "true or false")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildWeDriveFileSecureSettingRequest(*fileID, *text, *marginType, *showVisitorName, *showText)
	if err != nil {
		return err
	}
	return c.runWeDriveRequest(*dryRun, req, c.setWeDriveFileSecureSetting)
}

func weDriveFileIDCommand(c *wecomClient, name string, args []string, call func(any) error) error {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fileID := fs.String("fileid", "", "file ID")
	dryRun := fs.Bool("dry-run", false, "print request JSON without calling WeCom")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	req, err := buildWeDriveFileIDRequest(*fileID)
	if err != nil {
		return err
	}
	return c.runWeDriveRequest(*dryRun, req, call)
}

func buildWeDriveFileIDRequest(fileID string) (weDriveFileIDRequest, error) {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return weDriveFileIDRequest{}, errors.New("--fileid is required")
	}
	return weDriveFileIDRequest{FileID: fileID}, nil
}

func buildWeDriveFileListRequest(spaceID string, fatherID string, sortType int, start int, limit int) (weDriveFileListRequest, error) {
	spaceID = strings.TrimSpace(spaceID)
	fatherID = strings.TrimSpace(fatherID)
	if spaceID == "" {
		return weDriveFileListRequest{}, errors.New("--spaceid is required")
	}
	if fatherID == "" {
		return weDriveFileListRequest{}, errors.New("--fatherid is required")
	}
	if sortType == 0 {
		return weDriveFileListRequest{}, errors.New("--sort-type is required")
	}
	if start < 0 {
		return weDriveFileListRequest{}, errors.New("--start must be >= 0")
	}
	if limit < 1 || limit > 1000 {
		return weDriveFileListRequest{}, errors.New("--limit must be between 1 and 1000")
	}
	return weDriveFileListRequest{SpaceID: spaceID, FatherID: fatherID, SortType: sortType, Start: start, Limit: limit}, nil
}

func buildWeDriveFileCreateRequest(spaceID string, fatherID string, fileType int, fileName string) (weDriveFileCreateRequest, error) {
	spaceID = strings.TrimSpace(spaceID)
	fatherID = strings.TrimSpace(fatherID)
	fileName = strings.TrimSpace(fileName)
	if spaceID == "" {
		return weDriveFileCreateRequest{}, errors.New("--spaceid is required")
	}
	if fatherID == "" {
		return weDriveFileCreateRequest{}, errors.New("--fatherid is required")
	}
	if fileType != 1 && fileType != 3 && fileType != 4 {
		return weDriveFileCreateRequest{}, errors.New("--file-type must be 1, 3, or 4")
	}
	if fileName == "" {
		return weDriveFileCreateRequest{}, errors.New("--file-name is required")
	}
	return weDriveFileCreateRequest{SpaceID: spaceID, FatherID: fatherID, FileType: fileType, FileName: fileName}, nil
}

func buildWeDriveFileUploadRequest(spaceID string, fatherID string, selectedTicket string, fileName string, path string, base64Content string) (weDriveFileUploadRequest, error) {
	fileName = strings.TrimSpace(fileName)
	path = strings.TrimSpace(path)
	base64Content = strings.TrimSpace(base64Content)
	if (path == "") == (base64Content == "") {
		return weDriveFileUploadRequest{}, errors.New("exactly one of --path or --base64 is required")
	}
	if path != "" {
		raw, err := os.ReadFile(expandPath(path))
		if err != nil {
			return weDriveFileUploadRequest{}, fmt.Errorf("read --path: %w", err)
		}
		base64Content = base64.StdEncoding.EncodeToString(raw)
		if fileName == "" {
			fileName = filepath.Base(path)
		}
	}
	req, err := buildWeDriveUploadDestination(spaceID, fatherID, selectedTicket)
	if err != nil {
		return weDriveFileUploadRequest{}, err
	}
	if fileName == "" {
		return weDriveFileUploadRequest{}, errors.New("--file-name is required")
	}
	req.FileName = fileName
	req.FileBase64Content = base64Content
	return req, nil
}

func buildWeDriveUploadDestination(spaceID string, fatherID string, selectedTicket string) (weDriveFileUploadRequest, error) {
	spaceID = strings.TrimSpace(spaceID)
	fatherID = strings.TrimSpace(fatherID)
	selectedTicket = strings.TrimSpace(selectedTicket)
	hasSpace := spaceID != "" || fatherID != ""
	hasTicket := selectedTicket != ""
	if hasSpace == hasTicket {
		return weDriveFileUploadRequest{}, errors.New("pass exactly one of (--spaceid and --fatherid) or --selected-ticket")
	}
	if hasSpace && (spaceID == "" || fatherID == "") {
		return weDriveFileUploadRequest{}, errors.New("--spaceid and --fatherid must be passed together")
	}
	return weDriveFileUploadRequest{SpaceID: spaceID, FatherID: fatherID, SelectedTicket: selectedTicket}, nil
}

func buildWeDriveFileDownloadRequest(fileID string, selectedTicket string) (weDriveFileDownloadRequest, error) {
	fileID = strings.TrimSpace(fileID)
	selectedTicket = strings.TrimSpace(selectedTicket)
	if (fileID == "") == (selectedTicket == "") {
		return weDriveFileDownloadRequest{}, errors.New("exactly one of --fileid or --selected-ticket is required")
	}
	return weDriveFileDownloadRequest{FileID: fileID, SelectedTicket: selectedTicket}, nil
}

func buildWeDriveFileRenameRequest(fileID string, newName string) (weDriveFileRenameRequest, error) {
	idReq, err := buildWeDriveFileIDRequest(fileID)
	if err != nil {
		return weDriveFileRenameRequest{}, err
	}
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return weDriveFileRenameRequest{}, errors.New("--new-name is required")
	}
	return weDriveFileRenameRequest{FileID: idReq.FileID, NewName: newName}, nil
}

func buildWeDriveFileMoveRequest(fatherID string, replace string, fileIDs []string) (weDriveFileMoveRequest, error) {
	fatherID = strings.TrimSpace(fatherID)
	if fatherID == "" {
		return weDriveFileMoveRequest{}, errors.New("--fatherid is required")
	}
	replaceValue, err := parseOptionalBool(replace, "--replace")
	if err != nil {
		return weDriveFileMoveRequest{}, err
	}
	ids := uniqueStrings(fileIDs)
	if len(ids) == 0 {
		return weDriveFileMoveRequest{}, errors.New("--fileid is required")
	}
	return weDriveFileMoveRequest{FatherID: fatherID, Replace: replaceValue, FileID: ids}, nil
}

func buildWeDriveFileDeleteRequest(fileIDs []string) (weDriveFileDeleteRequest, error) {
	ids := uniqueStrings(fileIDs)
	if len(ids) == 0 {
		return weDriveFileDeleteRequest{}, errors.New("--fileid is required")
	}
	return weDriveFileDeleteRequest{FileID: ids}, nil
}

func buildWeDriveFileACLRequest(fileID string, authInfo []weDriveAuthInfo) (weDriveFileACLRequest, error) {
	idReq, err := buildWeDriveFileIDRequest(fileID)
	if err != nil {
		return weDriveFileACLRequest{}, err
	}
	if len(authInfo) == 0 {
		return weDriveFileACLRequest{}, errors.New("--member or --department is required")
	}
	return weDriveFileACLRequest{FileID: idReq.FileID, AuthInfo: authInfo}, nil
}

func buildWeDriveFileAuthInfo(members []string, departments []string, requireAuth bool) ([]weDriveAuthInfo, error) {
	var out []weDriveAuthInfo
	for _, raw := range members {
		value, auth, hasAuth, err := parseWeDriveAuthSpec(raw)
		if err != nil {
			return nil, err
		}
		if requireAuth && !hasAuth {
			return nil, fmt.Errorf("member %q requires auth", raw)
		}
		if requireAuth && auth != 1 {
			return nil, fmt.Errorf("member %q has invalid auth; file ACL auth must be 1", raw)
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
		if requireAuth && auth != 1 {
			return nil, fmt.Errorf("department %q has invalid auth; file ACL auth must be 1", raw)
		}
		id, err := parseUint32(value, "departmentid")
		if err != nil {
			return nil, err
		}
		out = append(out, weDriveAuthInfo{Type: 2, DepartmentID: id, Auth: auth})
	}
	return out, nil
}

func buildWeDriveFileSettingRequest(fileID string, authScope int, auth int) (weDriveFileSettingRequest, error) {
	idReq, err := buildWeDriveFileIDRequest(fileID)
	if err != nil {
		return weDriveFileSettingRequest{}, err
	}
	if authScope < 1 || authScope > 5 {
		return weDriveFileSettingRequest{}, errors.New("--auth-scope must be between 1 and 5")
	}
	if auth != 0 && auth != 1 && auth != 4 {
		return weDriveFileSettingRequest{}, errors.New("--auth must be 1 or 4")
	}
	return weDriveFileSettingRequest{FileID: idReq.FileID, AuthScope: authScope, Auth: auth}, nil
}

func buildWeDriveFileSecureSettingRequest(fileID string, text string, marginType int, showVisitorName string, showText string) (weDriveFileSecureSettingRequest, error) {
	idReq, err := buildWeDriveFileIDRequest(fileID)
	if err != nil {
		return weDriveFileSecureSettingRequest{}, err
	}
	visitorName, err := parseOptionalBool(showVisitorName, "--show-visitor-name")
	if err != nil {
		return weDriveFileSecureSettingRequest{}, err
	}
	textVisible, err := parseOptionalBool(showText, "--show-text")
	if err != nil {
		return weDriveFileSecureSettingRequest{}, err
	}
	if marginType != 0 && marginType != 1 && marginType != 2 {
		return weDriveFileSecureSettingRequest{}, errors.New("--margin-type must be 1 or 2")
	}
	return weDriveFileSecureSettingRequest{
		FileID: idReq.FileID,
		Watermark: weDriveWatermarkSetting{
			Text:            strings.TrimSpace(text),
			MarginType:      marginType,
			ShowVisitorName: visitorName,
			ShowText:        textVisible,
		},
	}, nil
}

func buildWeDriveFileUploadInitRequest(spaceID string, fatherID string, selectedTicket string, fileName string, path string, skipPushCard string) (weDriveFileUploadInitRequest, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return weDriveFileUploadInitRequest{}, errors.New("--path is required")
	}
	path = expandPath(path)
	info, err := os.Stat(path)
	if err != nil {
		return weDriveFileUploadInitRequest{}, fmt.Errorf("stat --path: %w", err)
	}
	if info.IsDir() {
		return weDriveFileUploadInitRequest{}, errors.New("--path must be a file")
	}
	if info.Size() < 0 {
		return weDriveFileUploadInitRequest{}, errors.New("file size is invalid")
	}
	if fileName = strings.TrimSpace(fileName); fileName == "" {
		fileName = filepath.Base(path)
	}
	dest, err := buildWeDriveUploadDestination(spaceID, fatherID, selectedTicket)
	if err != nil {
		return weDriveFileUploadInitRequest{}, err
	}
	skipPush, err := parseOptionalBool(skipPushCard, "--skip-push-card")
	if err != nil {
		return weDriveFileUploadInitRequest{}, err
	}
	blockSHA, err := weDriveCumulativeSHA1(path)
	if err != nil {
		return weDriveFileUploadInitRequest{}, err
	}
	return weDriveFileUploadInitRequest{
		SpaceID:        dest.SpaceID,
		FatherID:       dest.FatherID,
		SelectedTicket: dest.SelectedTicket,
		FileName:       fileName,
		Size:           uint64(info.Size()),
		BlockSHA:       blockSHA,
		SkipPushCard:   skipPush,
	}, nil
}

func weDriveCumulativeSHA1(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file for sha1: %w", err)
	}
	defer f.Close()
	h := sha1.New()
	marshaler, ok := h.(encoding.BinaryMarshaler)
	if !ok {
		return nil, errors.New("sha1 implementation cannot expose cumulative state")
	}
	buf := make([]byte, weDriveUploadBlockSize)
	var states []string
	for {
		n, readErr := io.ReadFull(f, buf)
		if readErr == io.ErrUnexpectedEOF || readErr == io.EOF {
			if n > 0 {
				_, _ = h.Write(buf[:n])
			}
			final := h.Sum(nil)
			finalHex := hex.EncodeToString(final)
			if readErr == io.EOF && n == 0 && len(states) > 0 {
				states[len(states)-1] = finalHex
			} else {
				states = append(states, finalHex)
			}
			return states, nil
		}
		if readErr != nil {
			return nil, fmt.Errorf("read file for sha1: %w", readErr)
		}
		_, _ = h.Write(buf[:n])
		state, err := marshaler.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("marshal sha1 state: %w", err)
		}
		states = append(states, hex.EncodeToString(state[4:24]))
	}
}

func (c *wecomClient) uploadWeDriveFileInChunks(req weDriveFileUploadInitRequest, path string) error {
	var initResp weDriveFileUploadInitResponse
	if err := c.postWeDriveJSON("/cgi-bin/wedrive/file_upload_init", req, &initResp); err != nil {
		return err
	}
	if initResp.HitExist {
		return printPrettyJSON(initResp)
	}
	if initResp.UploadKey == "" {
		return errors.New("file_upload_init response did not include upload_key")
	}
	path = expandPath(path)
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open --path: %w", err)
	}
	defer f.Close()
	buf := make([]byte, weDriveUploadBlockSize)
	for index := 1; ; index++ {
		n, readErr := io.ReadFull(f, buf)
		if readErr == io.EOF {
			break
		}
		if readErr != nil && readErr != io.ErrUnexpectedEOF {
			return fmt.Errorf("read chunk %d: %w", index, readErr)
		}
		partReq := weDriveFileUploadPartRequest{
			UploadKey:         initResp.UploadKey,
			Index:             index,
			FileBase64Content: base64.StdEncoding.EncodeToString(buf[:n]),
		}
		var partResp apiErrorResponse
		if err := c.postWeDriveJSON("/cgi-bin/wedrive/file_upload_part", partReq, &partResp); err != nil {
			return fmt.Errorf("upload chunk %d: %w", index, err)
		}
		if readErr == io.ErrUnexpectedEOF {
			break
		}
	}
	finishReq := weDriveFileUploadFinishRequest{UploadKey: initResp.UploadKey}
	return c.postWeDriveJSON("/cgi-bin/wedrive/file_upload_finish", finishReq, nil)
}

func (c *wecomClient) postWeDriveJSON(path string, req any, out any) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}
	rawBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}
	u := c.cfg.BaseURL + path + "?access_token=" + url.QueryEscape(token)
	httpReq, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(rawBody))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("POST %s returned HTTP %d: %s", path, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var apiErr apiErrorResponse
	if json.Unmarshal(raw, &apiErr) == nil && apiErr.ErrCode != 0 {
		return fmt.Errorf("WeCom returned errcode %d: %s", apiErr.ErrCode, apiErr.ErrMsg)
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
		return nil
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		fmt.Println("{}")
		return nil
	}
	var formatted bytes.Buffer
	if json.Valid(raw) && json.Indent(&formatted, raw, "", "  ") == nil {
		fmt.Println(formatted.String())
		return nil
	}
	fmt.Println(string(raw))
	return nil
}

func (c *wecomClient) listWeDriveFiles(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/file_list", req)
}

func (c *wecomClient) getWeDriveFileInfo(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/file_info", req)
}

func (c *wecomClient) createWeDriveFile(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/file_create", req)
}

func (c *wecomClient) uploadWeDriveFile(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/file_upload", req)
}

func (c *wecomClient) downloadWeDriveFile(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/file_download", req)
}

func (c *wecomClient) renameWeDriveFile(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/file_rename", req)
}

func (c *wecomClient) moveWeDriveFile(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/file_move", req)
}

func (c *wecomClient) deleteWeDriveFile(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/file_delete", req)
}

func (c *wecomClient) shareWeDriveFile(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/file_share", req)
}

func (c *wecomClient) getWeDriveFilePermission(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/get_file_permission", req)
}

func (c *wecomClient) addWeDriveFileACL(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/file_acl_add", req)
}

func (c *wecomClient) removeWeDriveFileACL(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/file_acl_del", req)
}

func (c *wecomClient) setWeDriveFileSetting(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/file_setting", req)
}

func (c *wecomClient) setWeDriveFileSecureSetting(req any) error {
	return c.postWeDrive("/cgi-bin/wedrive/file_secure_setting", req)
}

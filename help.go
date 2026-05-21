package main

import "fmt"

func printUsage() {
	fmt.Print(`wecom-cli operates Tencent WeCom APIs.

Usage:
  wecom-cli [global flags] <command> <subcommand> [flags]

Global flags:
  --corpid        WeCom enterprise ID. Defaults to WECOM_CORP_ID
  --corpsecret    WeCom app secret. Defaults to WECOM_CORP_SECRET
  --token-cache   access_token cache file. Defaults to ~/.wecom-cli/access_tokens.json

Commands:
  calendar   Create and manage calendars
  schedule   Create and manage schedules
  meeting    Create and manage reserved meetings
  wedrive    Manage WeDrive spaces

Run "<command> help" for command-specific details.
`)
}

func printCalendarUsage() {
	fmt.Print(`Calendar commands:
  wecom-cli calendar create --summary "Project Launch" --color "#2F7DFF" [flags]
  wecom-cli calendar update --cal-id CAL_ID --summary "Project Launch" --color "#2F7DFF" [flags]
  wecom-cli calendar get --cal-id CAL_ID [--cal-id CAL_ID]
  wecom-cli calendar delete --cal-id CAL_ID

Create flags:
  --summary             Calendar title. Required
  --color               RGB color such as #2F7DFF. Required unless --corp-calendar
  --description         Calendar description
  --admin               Calendar admin userid. Repeatable
  --share               userid[:permission]. Repeatable. permission: 1=view, 3=free/busy
  --default             Set as app default calendar
  --public              Create public calendar
  --public-user         Public-range userid. Repeatable
  --public-party        Public-range department ID. Repeatable
  --corp-calendar       Create all-staff calendar
  --agentid             Authorized app agentid
  --dry-run             Print request JSON without calling WeCom

Update flags:
  --cal-id              Calendar ID. Required
  --summary             Calendar title. Required
  --color               RGB color such as #2F7DFF. Required
  --description         Calendar description
  --admin               Calendar admin userid. Repeatable
  --share               userid[:permission]. Repeatable. permission: 1=view, 3=free/busy
  --public-user         Public-range userid. Repeatable
  --public-party        Public-range department ID. Repeatable
  --skip-public-range   Do not update public subscription range
  --dry-run             Print request JSON without calling WeCom

Get flags:
  --cal-id              Calendar ID. Repeatable. Required
  --dry-run             Print request JSON without calling WeCom

Delete flags:
  --cal-id              Calendar ID. Required
  --dry-run             Print request JSON without calling WeCom
`)
}

func printScheduleUsage() {
	fmt.Print(`Schedule commands:
  wecom-cli schedule create --start TIME --end TIME [flags]
  wecom-cli schedule update --schedule-id ID --start TIME --end TIME [flags]
  wecom-cli schedule get --schedule-id ID [--schedule-id ID]
  wecom-cli schedule list --cal-id CAL_ID [--offset 0] [--limit 500]
  wecom-cli schedule delete --schedule-id ID [--op-mode 1 --op-start-time TIME]
  wecom-cli schedule add-attendees --schedule-id ID --attendee USERID
  wecom-cli schedule remove-attendees --schedule-id ID --attendee USERID

Create/update flags:
  --start                 Start time. Unix seconds or RFC3339. Required
  --end                   End time. Unix seconds or RFC3339. Required
  --cal-id                Calendar ID. Create only
  --schedule-id           Schedule ID. Update only
  --agentid               Authorized app agentid. Create only
  --summary               Schedule title
  --description           Schedule description
  --location              Schedule location
  --whole-day             Mark as whole-day schedule
  --admin                 Admin userid. Repeatable
  --attendee              Attendee userid. Repeatable
  --remind                Enable reminder
  --remind-before         Seconds before event reminder
  --remind-time-diff      Reminder diff from start time. Repeatable
  --repeat                Enable repeat schedule
  --repeat-type           0 daily, 1 weekly, 2 monthly, 5 yearly, 7 workday
  --repeat-until          Repeat end time. Unix seconds or RFC3339
  --custom-repeat         Enable custom repeat
  --repeat-interval       Custom repeat interval
  --repeat-day-of-week    Custom repeat weekday. Repeatable
  --repeat-day-of-month   Custom repeat month day. Repeatable
  --timezone              UTC offset, -12 to 12
  --skip-attendees        Do not update attendees. Update only
  --op-mode               Repeat operation mode. 0 all, 1 current, 2 future
  --op-start-time         Repeat occurrence start time for op-mode 1 or 2
  --dry-run               Print request JSON without calling WeCom

Get/list/delete/attendee flags:
  --schedule-id           Schedule ID. Repeatable for get
  --cal-id                Calendar ID for list
  --offset                List pagination offset
  --limit                 List pagination limit, 1-1000
  --attendee              Attendee userid. Repeatable
  --op-mode               Delete repeat operation mode. 0 all, 1 current, 2 future
  --op-start-time         Delete repeat occurrence start time for op-mode 1 or 2
  --dry-run               Print request JSON without calling WeCom
`)
}

func printMeetingUsage() {
	fmt.Print(`Meeting commands:
  wecom-cli meeting create --admin-userid USERID --title TITLE --start TIME --duration SECONDS [flags]
  wecom-cli meeting update --meeting-id ID [flags]
  wecom-cli meeting get --meeting-id ID
  wecom-cli meeting list --userid USERID [--begin TIME] [--end TIME] [--cursor CURSOR] [--limit 100]
  wecom-cli meeting cancel --meeting-id ID

Create/update flags:
  --meeting-id             Meeting ID. Update/get/cancel only
  --admin-userid           Meeting admin userid. Create required
  --title                  Meeting title. Create required
  --start                  Meeting start time. Unix seconds or RFC3339
  --duration               Meeting duration in seconds, 300-86399
  --description            Meeting description
  --location               Meeting location
  --agentid                Authorized app agentid
  --cal-id                 Calendar ID
  --invitee                Invitee userid. Repeatable
  --remind-scope           1 none, 2 hosts, 3 all, 4 ring-users
  --password               Meeting password, 4-6 digits
  --waiting-room           true or false
  --allow-enter-before-host true or false
  --enter-mute             0 off, 1 on, 2 auto
  --screen-watermark       true or false
  --host                   Host userid. Repeatable
  --ring-user              Ring userid. Repeatable
  --repeat                 Enable repeat meeting
  --repeat-type            0 daily, 1 weekly, 2 monthly, 7 workday
  --repeat-until           Repeat end time. Unix seconds or RFC3339
  --repeat-interval        Repeat interval
  --remind-before          Meeting reminder before start in seconds. Repeatable
  --dry-run                Print request JSON without calling WeCom

List flags:
  --userid                 Member userid
  --begin                  Begin time. Unix seconds or RFC3339
  --end                    End time. Unix seconds or RFC3339
  --cursor                 Pagination cursor
  --limit                  Page size, 1-100
  --dry-run                Print request JSON without calling WeCom
`)
}

func printWeDriveUsage() {
	fmt.Print(`WeDrive commands:
  wecom-cli wedrive space <subcommand> [flags]
  wecom-cli wedrive file <subcommand> [flags]

Run "wecom-cli wedrive space help" for details.
Run "wecom-cli wedrive file help" for details.
`)
}

func printWeDriveSpaceUsage() {
	fmt.Print(`WeDrive space commands:
  wecom-cli wedrive space create --space-name NAME [--member USERID:AUTH] [--department ID:AUTH]
  wecom-cli wedrive space info --spaceid SPACEID
  wecom-cli wedrive space new-info --spaceid SPACEID
  wecom-cli wedrive space rename --spaceid SPACEID --space-name NAME
  wecom-cli wedrive space dismiss --spaceid SPACEID
  wecom-cli wedrive space share --spaceid SPACEID
  wecom-cli wedrive space acl-add --spaceid SPACEID --member USERID:AUTH
  wecom-cli wedrive space acl-del --spaceid SPACEID --member USERID
  wecom-cli wedrive space setting --spaceid SPACEID [flags]

Auth specs:
  --member USERID:AUTH        Member auth. Auth: 1 download, 4 preview, 7 space admin
  --department ID:AUTH        Department auth. Auth: 1 download
  acl-del accepts --member USERID and --department ID without auth

Setting flags:
  --enable-watermark true|false
  --share-url-no-approve true|false
  --share-url-no-approve-default-auth 1|2|4|5|200
  --enable-confidential-mode true|false
  --default-file-scope 1|2
  --ban-share-external true|false
  --dry-run
`)
}

func printWeDriveFileUsage() {
	fmt.Print(`WeDrive file commands:
  wecom-cli wedrive file list --spaceid SPACEID --fatherid FATHERID
  wecom-cli wedrive file info --fileid FILEID
  wecom-cli wedrive file create --spaceid SPACEID --fatherid FATHERID --file-type 1 --file-name NAME
  wecom-cli wedrive file upload --spaceid SPACEID --fatherid FATHERID --path ./report.pdf
  wecom-cli wedrive file upload-chunk --spaceid SPACEID --fatherid FATHERID --path ./large.zip
  wecom-cli wedrive file download --fileid FILEID
  wecom-cli wedrive file rename --fileid FILEID --new-name NAME
  wecom-cli wedrive file move --fatherid FOLDER_ID --fileid FILEID [--replace true]
  wecom-cli wedrive file delete --fileid FILEID
  wecom-cli wedrive file share --fileid FILEID
  wecom-cli wedrive file permission --fileid FILEID
  wecom-cli wedrive file acl-add --fileid FILEID --member USERID:1
  wecom-cli wedrive file acl-del --fileid FILEID --member USERID
  wecom-cli wedrive file setting --fileid FILEID --auth-scope 2 --auth 1
  wecom-cli wedrive file secure-setting --fileid FILEID [flags]

File flags:
  --spaceid              Space ID
  --fatherid             Parent folder file ID. Root uses the space ID
  --fileid               File ID. Repeatable for move/delete
  --file-type            1 folder, 3 doc, 4 sheet
  --file-name            File name. Upload defaults to --path basename
  --path                 Local file path for upload/upload-chunk
  --base64               Base64 file content for normal upload
  --selected-ticket      selectedTicket from WeDrive/file picker JSAPI
  --sort-type            Sort type for list
  --start                Pagination start for list
  --limit                Pagination limit for list, 1-1000
  --replace              Move overwrite flag, true or false
  --auth-scope           1 specified users, 2 corp internal, 3 corp external, 4 internal approval, 5 external approval
  --auth                 File share/member auth. 1 browse/download, 4 preview
  --member               userid[:auth]. Repeatable
  --department           departmentid[:auth]. Repeatable
  --watermark-text       Watermark text for secure-setting
  --margin-type          1 low-density, 2 high-density
  --show-visitor-name    true or false
  --show-text            true or false
  --skip-push-card       Chunk upload completion card flag, true or false
  --dry-run
`)
}

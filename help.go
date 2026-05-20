package main

import "fmt"

func printUsage() {
	fmt.Print(`wecom-cli operates Tencent WeCom APIs.

Usage:
  wecom-cli [global flags] <command> <subcommand> [flags]

Global flags:
  --base-url      WeCom API base URL. Defaults to https://qyapi.weixin.qq.com
  --corpid        WeCom enterprise ID. Defaults to WECOM_CORP_ID
  --corpsecret    WeCom app secret. Defaults to WECOM_CORP_SECRET
  --agent-id      WeCom agent ID. Defaults to WECOM_AGENT_ID
  --token-cache   access_token cache file. Defaults to ~/.wecom-cli/access_tokens.json
  --agw-cli       agw-cli path for employee-name lookups. Defaults to AGW_CLI

Commands:
  calendar   Create and manage calendars
  schedule   Create and manage schedules

Run "wecom-cli calendar help" or "wecom-cli schedule help" for details.
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
  --admin-name          Admin employee name to resolve through agw-cli. Repeatable
  --share               userid[:permission]. Repeatable. permission: 1=view, 3=free/busy
  --share-name          employee-name[:permission]. Repeatable
  --default             Set as app default calendar
  --public              Create public calendar
  --public-user         Public-range userid. Repeatable
  --public-user-name    Public-range employee name. Repeatable
  --public-party        Public-range department ID. Repeatable
  --corp-calendar       Create all-staff calendar
  --dry-run             Print request JSON without calling WeCom

Update flags:
  --cal-id              Calendar ID. Required
  --summary             Calendar title. Required
  --color               RGB color such as #2F7DFF. Required
  --description         Calendar description
  --admin               Calendar admin userid. Repeatable
  --admin-name          Admin employee name to resolve through agw-cli. Repeatable
  --share               userid[:permission]. Repeatable. permission: 1=view, 3=free/busy
  --share-name          employee-name[:permission]. Repeatable
  --public-user         Public-range userid. Repeatable
  --public-user-name    Public-range employee name. Repeatable
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
  --summary               Schedule title
  --description           Schedule description
  --location              Schedule location
  --whole-day             Mark as whole-day schedule
  --admin                 Admin userid. Repeatable
  --admin-name            Admin employee name to resolve through agw-cli. Repeatable
  --attendee              Attendee userid. Repeatable
  --attendee-name         Attendee employee name to resolve through agw-cli. Repeatable
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
  --attendee-name         Attendee employee name to resolve through agw-cli. Repeatable
  --op-mode               Delete repeat operation mode. 0 all, 1 current, 2 future
  --op-start-time         Delete repeat occurrence start time for op-mode 1 or 2
  --dry-run               Print request JSON without calling WeCom
`)
}

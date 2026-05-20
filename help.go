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

Run "wecom-cli calendar help" for details.
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

# wecom-cli

CLI for agents to operate Tencent WeCom. The first implemented capability is
creating calendars through the official `oa/calendar/add` API.

## Build

```bash
go build -o wecom-cli .
```

## Configuration

The CLI loads `.env` from the current working directory when present. Real
environment variables and command-line flags take precedence.

Required for WeCom API calls:

```env
WECOM_CORP_ID=wwxxxxxxxxxxxxxxxx
WECOM_CORP_SECRET=your_app_secret
```

Optional:

```env
WECOM_AGENT_ID=1000014
WECOM_BASE_URL=https://qyapi.weixin.qq.com
WECOM_TOKEN_CACHE=~/.wecom-cli/access_tokens.json
AGW_CLI=/Users/suwenhao/agw-cli/agw-cli
```

`WECOM_TOKEN_CACHE` defaults to `~/.wecom-cli/access_tokens.json`. Access tokens
are cached per `corpid + secret` and refreshed before expiry.

## Calendar Commands

Create a calendar with raw WeCom user IDs:

```bash
./wecom-cli calendar create \
  --summary "Project Launch" \
  --color "#2F7DFF" \
  --description "Launch coordination calendar" \
  --admin alice \
  --share bob:1 \
  --share charlie:3
```

Resolve employee names through `agw-cli`:

```bash
./wecom-cli calendar create \
  --summary "Project Launch" \
  --color "#2F7DFF" \
  --admin-name "张三" \
  --share-name "李四:1"
```

Useful flags:

```text
--summary               Calendar title, required
--color                 RGB color such as #2F7DFF, required unless --corp-calendar
--description           Calendar description
--admin                 Calendar admin userid, repeatable
--admin-name            Admin employee name to resolve via agw-cli, repeatable
--share                 userid[:permission], repeatable. permission: 1=view, 3=free/busy
--share-name            employee-name[:permission], repeatable
--default               Set as app default calendar
--public                Create public calendar
--public-user           Public-range userid, repeatable
--public-user-name      Public-range employee name, repeatable
--public-party          Public-range department id, repeatable
--corp-calendar         Create all-staff calendar
--agent-id              WeCom agentid. Defaults to WECOM_AGENT_ID when present
--dry-run               Print request JSON without calling WeCom
```

Update a calendar. The official API is overwrite-style, so pass the full values
you want to keep for required fields and any list fields you include:

```bash
./wecom-cli calendar update \
  --cal-id wcdV_NCwAAfRaN-5hxmoypLk7GeoeOCA \
  --summary "Project Launch Updated" \
  --color "#FF3030" \
  --description "Updated launch coordination calendar" \
  --admin alice \
  --share bob:1
```

If the calendar has a public subscription range and you do not want to overwrite
it during update:

```bash
./wecom-cli calendar update \
  --cal-id wcdV_NCwAAfRaN-5hxmoypLk7GeoeOCA \
  --summary "Project Launch Updated" \
  --color "#FF3030" \
  --skip-public-range
```

Get one or more calendars:

```bash
./wecom-cli calendar get \
  --cal-id wcdV_NCwAAfRaN-5hxmoypLk7GeoeOCA
```

Delete a calendar:

```bash
./wecom-cli calendar delete \
  --cal-id wcdV_NCwAAfRaN-5hxmoypLk7GeoeOCA
```

## Schedule Commands

Create a schedule. Times accept Unix seconds or RFC3339:

```bash
./wecom-cli schedule create \
  --cal-id wcdV_NCwAAfRaN-5hxmoypLk7GeoeOCA \
  --summary "Project Review" \
  --start 2026-05-20T15:00:00+08:00 \
  --end 2026-05-20T16:00:00+08:00 \
  --attendee 029235 \
  --remind-time-diff -300
```

Update a schedule. The official update API is overwrite-style; use
`--skip-attendees` when you do not want to overwrite the attendee list:

```bash
./wecom-cli schedule update \
  --schedule-id SCHEDULE_ID \
  --summary "Project Review Updated" \
  --start 2026-05-20T15:00:00+08:00 \
  --end 2026-05-20T16:30:00+08:00 \
  --skip-attendees
```

Get schedule details or list schedules under a calendar:

```bash
./wecom-cli schedule get --schedule-id SCHEDULE_ID

./wecom-cli schedule list \
  --cal-id wcdV_NCwAAfRaN-5hxmoypLk7GeoeOCA \
  --offset 0 \
  --limit 100
```

Cancel a schedule:

```bash
./wecom-cli schedule delete --schedule-id SCHEDULE_ID
```

For repeated schedules, pass the operation mode and occurrence start time:

```bash
./wecom-cli schedule delete \
  --schedule-id SCHEDULE_ID \
  --op-mode 1 \
  --op-start-time 2026-05-20T15:00:00+08:00
```

Add or remove attendees without overwriting the whole attendee list:

```bash
./wecom-cli schedule add-attendees \
  --schedule-id SCHEDULE_ID \
  --attendee 029235

./wecom-cli schedule remove-attendees \
  --schedule-id SCHEDULE_ID \
  --attendee 029235
```

## Meeting Commands

Create a reserved meeting:

```bash
./wecom-cli meeting create \
  --admin-userid 029235 \
  --title "Project Review" \
  --start 2026-05-20T15:00:00+08:00 \
  --duration 1800 \
  --invitee 029235 \
  --location "Room 1005"
```

Optional meeting settings can be passed as flags:

```bash
./wecom-cli meeting create \
  --admin-userid 029235 \
  --title "Project Review" \
  --start 2026-05-20T15:00:00+08:00 \
  --duration 1800 \
  --invitee 029235 \
  --waiting-room false \
  --allow-enter-before-host true \
  --enter-mute 2 \
  --remind-before 900
```

Update, fetch, list, or cancel meetings:

```bash
./wecom-cli meeting update \
  --meeting-id MEETING_ID \
  --title "Updated Project Review" \
  --start 2026-05-20T16:00:00+08:00 \
  --duration 3600

./wecom-cli meeting get --meeting-id MEETING_ID

./wecom-cli meeting list \
  --userid 029235 \
  --begin 2026-05-01T00:00:00+08:00 \
  --end 2026-05-31T23:59:59+08:00

./wecom-cli meeting cancel --meeting-id MEETING_ID
```

## WeDrive Space Commands

Create a WeDrive space and grant initial permissions:

```bash
./wecom-cli wedrive space create \
  --space-name "Project Space" \
  --member 029235:7
```

Manage space metadata:

```bash
./wecom-cli wedrive space info --spaceid SPACEID
./wecom-cli wedrive space new-info --spaceid SPACEID
./wecom-cli wedrive space rename --spaceid SPACEID --space-name "Renamed Space"
./wecom-cli wedrive space share --spaceid SPACEID
./wecom-cli wedrive space dismiss --spaceid SPACEID
```

Manage members or departments:

```bash
./wecom-cli wedrive space acl-add \
  --spaceid SPACEID \
  --member 029235:1

./wecom-cli wedrive space acl-del \
  --spaceid SPACEID \
  --member 029235
```

Update security settings:

```bash
./wecom-cli wedrive space setting \
  --spaceid SPACEID \
  --enable-watermark false \
  --share-url-no-approve true \
  --share-url-no-approve-default-auth 1 \
  --default-file-scope 2
```

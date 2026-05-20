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

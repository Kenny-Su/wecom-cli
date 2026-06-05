# wecom-cli

CLI for operating Tencent WeCom APIs from scripts and agent workflows.

`wecom-cli` does not call Tencent WeCom directly. It sends requests through a
gateway using bearer-token authentication. The gateway is responsible for
injecting or transforming the Tencent WeCom `access_token`.

The CLI also integrates with AGW admin APIs to store resources created by WeCom
operations and to query employee-WeCom user mappings.

## Build

```bash
go build -o wecom-cli .
```

## Configuration

The CLI loads `.env` from the current working directory when present. When run
as a packaged skill binary, it also loads `.env` next to the nearest ancestor
`SKILL.md`. Real environment variables and command-line flags take precedence.

Required:

```env
WECOM_GATEWAY_BASE_URL=https://gateway.example.com/wecom
AGW_GATEWAY_BASE_URL=https://gateway.example.com
CLI_IDENTITY_FILE=/path/to/cli-identity.env
```

`CLI_IDENTITY_FILE` must point to a JSON file containing:

```json
{"ACCESS_TOKEN":"your_gateway_token"}
```

`WECOM_GATEWAY_BASE_URL` is used for WeCom API paths such as
`/cgi-bin/oa/calendar/add`.

`AGW_GATEWAY_BASE_URL` is used for resource storage and employee-WeCom mapping
lookups. If it is not set, the CLI derives it from `WECOM_GATEWAY_BASE_URL` by
removing a trailing `/wecom` path segment.

All gateway and AGW requests send:

```http
Authorization: Bearer <ACCESS_TOKEN>
```

## Commands

WeCom operation commands:

```bash
wecom-cli calendar help
wecom-cli schedule help
wecom-cli meeting help
wecom-cli wedrive help
```

AGW resource commands:

```bash
wecom-cli resources list
wecom-cli resources get --id 1
wecom-cli resources add --resource-type calendar --platform-field cal_id --external-id CAL_ID
```

Employee-WeCom mapping commands:

```bash
wecom-cli users get-by-name --user-name "Zhang San"
wecom-cli users get-by-qw-user --qw-userid qw-1
wecom-cli users get-by-staff-id --staff-id staff-1
wecom-cli users list
```

## Automatic Resource Storage

Successful create/upload commands automatically store returned WeCom IDs as AGW
user-agent resources:

- calendars
- schedules
- meetings
- WeDrive spaces
- WeDrive files

Use `wecom-cli resources list` to inspect stored resources.

Most mutating WeCom commands support `--dry-run` to print request JSON without
calling WeCom or storing resources.

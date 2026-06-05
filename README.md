# wecom-cli

CLI for operating Tencent WeCom APIs from scripts and agent workflows.

The command surface follows the official WeCom API documentation.

## Build

Build with the standard Go toolchain from the repository root.

## Configuration

The CLI loads `.env` from the current working directory when present. When run
as a packaged skill binary, it also loads `.env` next to the nearest ancestor
`SKILL.md`. Real environment variables and command-line flags take precedence.

Required for WeCom API calls:

```env
WECOM_GATEWAY_BASE_URL=https://gateway.example.com/wecom
AGW_GATEWAY_BASE_URL=https://gateway.example.com
CLI_IDENTITY_FILE=/path/to/cli-identity.env
```

Requests are sent to `WECOM_GATEWAY_BASE_URL` with the original WeCom API path.
The CLI reads `CLI_IDENTITY_FILE` from the system environment first, falls back to `.env` when needed, then reads `ACCESS_TOKEN` from that file
and sends it as `Authorization: Bearer <token>`. The relay gateway is
responsible for adding or transforming the WeCom `access_token`.

`AGW_GATEWAY_BASE_URL` is used for resource storage and employee-WeCom mapping
lookups. If it is not set, the CLI derives it from `WECOM_GATEWAY_BASE_URL` by
removing a trailing `/wecom` path segment.

## Usage

Use the built-in help as the source of truth for commands, flags, and examples.
Each command group also has command-specific help.

Most mutating commands support `--dry-run` to print the request JSON without
calling WeCom.

Successful create/upload commands for calendars, schedules, meetings, WeDrive
spaces, and WeDrive files automatically store the returned WeCom ID as an AGW
user-agent resource. Use `wecom-cli resources list` to check stored resources
and `wecom-cli users get-by-name --user-name NAME` or related `users` commands
to query employee WeCom user mappings.

## References

Official API notes and copied reference material live under `references/`.

# wecom-cli

CLI for operating Tencent WeCom APIs from scripts and agent workflows.

The command surface follows the official WeCom API documentation.

## Build

Build with the standard Go toolchain from the repository root.

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
WECOM_BASE_URL=https://qyapi.weixin.qq.com
WECOM_TOKEN_CACHE=~/.wecom-cli/access_tokens.json
WECOM_RESOURCE_TABLE=~/.wecom-cli/resources.json
```

`WECOM_BASE_URL` defaults to `https://qyapi.weixin.qq.com`.

`WECOM_TOKEN_CACHE` defaults to `~/.wecom-cli/access_tokens.json`. Access tokens
are cached per `corpid + secret` and refreshed before expiry.

`WECOM_RESOURCE_TABLE` defaults to `~/.wecom-cli/resources.json`. The CLI writes
successful create responses there so agents can look up API IDs for resources
they created earlier.

## Usage

Use the built-in help as the source of truth for commands, flags, and examples.
Each command group also has command-specific help.

Most mutating commands support `--dry-run` to print the request JSON without
calling WeCom.

## Created Resource Table

WeCom APIs usually require resource IDs for follow-up operations. For example,
a schedule cannot be fetched by name; it must be fetched by `schedule_id`.

After a successful create call, the CLI records the returned ID and useful
metadata in the resource table. Tracked resources include:

- `calendar`
- `schedule`
- `meeting`
- `wedrive_space`
- `wedrive_file`

Use the `resources` help output to see how to inspect the table path, list
tracked records, filter by type, print JSON, or override the table path for
isolated tests.

The table keeps historical records. Deleting or cancelling a remote resource
does not remove the local record.

Sensitive or large request fields such as meeting passwords, selected tickets,
and uploaded base64 file content are redacted before being stored.

## References

Official API notes and copied reference material live under `references/`.

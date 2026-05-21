# wecom-cli

CLI for operating Tencent WeCom APIs from scripts and agent workflows.

The command surface follows the official WeCom API documentation.

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
WECOM_TOKEN_CACHE=~/.wecom-cli/access_tokens.json
```

`WECOM_TOKEN_CACHE` defaults to `~/.wecom-cli/access_tokens.json`. Access tokens
are cached per `corpid + secret` and refreshed before expiry.

## Usage

Use the built-in help as the source of truth for commands, flags, and examples:

```bash
./wecom-cli help
./wecom-cli calendar help
./wecom-cli schedule help
./wecom-cli meeting help
./wecom-cli wedrive help
./wecom-cli wedrive space help
./wecom-cli wedrive file help
```

Most mutating commands support `--dry-run` to print the request JSON without
calling WeCom.

## References

Official API notes and copied reference material live under `references/`.

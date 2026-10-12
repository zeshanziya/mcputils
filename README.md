# mcputils

A CLI that manages MCP (Model Context Protocol) server configurations for
Claude Code, backed by a YAML catalog and named profiles. Enabling or
disabling a profile toggles all of its servers at once; individual servers
can also be toggled directly. mcputils only edits config state — it does not
start, stop, or otherwise manage MCP server processes.

## Install

### Homebrew (macOS/Linux)

```sh
brew tap zeshanziya/mcputils
brew install mcputils
```

### Download a prebuilt binary

Grab the archive for your platform from the [latest release](https://github.com/zeshanziya/mcputils/releases/latest), then:

```sh
curl -sL https://github.com/zeshanziya/mcputils/releases/latest/download/mcputils_darwin_arm64.tar.gz | tar xz
sudo mv mcputils /usr/local/bin/
```

Replace `darwin_arm64` with your platform:

| OS | Arch | Archive |
|----|------|---------|
| macOS (Apple Silicon) | arm64 | `mcputils_darwin_arm64.tar.gz` |
| macOS (Intel) | amd64 | `mcputils_darwin_amd64.tar.gz` |
| Linux | amd64 | `mcputils_linux_amd64.tar.gz` |
| Linux | arm64 | `mcputils_linux_arm64.tar.gz` |
| Windows | amd64 | `mcputils_windows_amd64.zip` |
| Windows | arm64 | `mcputils_windows_arm64.zip` |

A `checksums.txt` is attached to every release if you want to verify the
download.

### With Go

```sh
go install github.com/zeshanziya/mcputils@latest
```

## Usage

mcputils operates on one of two scopes, chosen with `--global` or
`--project`:

- `--global` — applies to Claude Code's user-level config (`~/.claude.json`,
  top-level `mcpServers`).
- `--project` — applies to the current project (auto-detected by walking up
  from the current directory to find a `.mcp.json`), or pass an explicit
  path: `--project /path/to/project`.
- With neither flag, mcputils auto-detects project scope from the current
  directory; if no `.mcp.json` is found anywhere up the tree, it errors and
  asks you to pass `--global` or `--project`.

### Catalog and profiles

Servers live in a YAML catalog (`~/.config/mcputils/servers.yaml` for global
scope, `.mcputils/servers.yaml` for project scope) and are grouped into
named profiles.

```sh
# create a profile
mcputils profile create work

# add/remove servers from a profile
mcputils profile add-server work fs
mcputils profile remove-server work fs
```

The catalog file itself is hand-edited for now (to add new server
definitions); see the format below.

### Enable / disable

```sh
# enable or disable every server in a profile
mcputils --global profile enable work
mcputils --global profile disable work

# enable or disable one server directly, independent of any profile
mcputils --project server enable fs
mcputils --project server disable fs
```

Multiple profiles can be active at once. Disabling a profile only turns off
servers that aren't also listed in another currently-active profile.
Profile `enable`/`disable` act strictly on their own member list and are
never blocked by a prior individual `server enable`/`disable` call.

### Listing and status

```sh
mcputils list servers      # catalog servers with enabled/disabled state
mcputils list profiles     # profiles with active/inactive state
mcputils status            # combined view: each profile, its servers, effective state
```

### History

Every effective enable/disable change (not every CLI invocation — only
actual state transitions) is appended to a single, global, append-only log
at `~/.config/mcputils/history.jsonl`:

```sh
mcputils log                # print the history, newest first
```

## Catalog format

```yaml
servers:
  - name: fs
    command: fs-server
    args: ["--root", "/"]
    env:
      API_KEY: "$API_KEY"   # passthrough only — mcputils never resolves env values
profiles:
  - name: work
    servers: [fs]
```

Catalog files are stored at `0600` and their parent directories at `0700`
since they can hold literal secret values in `env` — only the owning user
can read them.

## Supported clients

v1 supports Claude Code only, applied to a single file, `~/.claude.json`.
The enable/disable mechanism differs by scope:

- **Project scope**: toggles the server's name in
  `projects[<projectRoot>].disabledMcpServers`; the server's own config entry
  is left untouched.
- **Global scope**: there is no top-level disabled-array in real Claude Code,
  so disabling removes the server's entry from the top-level `mcpServers`
  map, and enabling re-creates it from the catalog and sweeps the server's
  name out of every project's `disabledMcpServers` array (so "enabled
  globally" means enabled everywhere, not just available again).

Config writes are atomic (temp file + rename) and preserve unrelated keys.

The client integration lives behind a `ClientAdapter` interface, so other
clients (Claude Desktop, Cursor, etc.) could be added later without
reworking the core profile/catalog logic — not implemented in v1.

## Non-goals

- No server process/container lifecycle management — config toggling only.
- No secret resolution/substitution — `env` values pass through as-is.
- No concurrent-process safety — single-user CLI; simultaneous invocations
  racing on the same config file may lose an update.


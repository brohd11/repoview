# repoview

A manifest-free git repo-status viewer for a folder full of projects. Point it at a directory
and it scans for every nested git checkout (fresh each run) and shows each one's status — branch,
uncommitted changes, ahead/behind — in a single TUI list, with fetch/pull/push/commit a keypress
away.

```
repoview [dir] [depth]   # TUI; dir defaults to the current directory, depth to 1 (that dir only)
  repoview 4             # a bare integer is the depth — scan 4 levels deep
  repoview /path 3       # scan /path, 3 levels deep
  -d, --depth N          # depth can also be given as a flag (default 1)
```

Keys: **enter** (or **v**) opens the highlighted repo's git menu (status/fetch/pull/push/commit),
**t** opens a terminal at that repo's directory, **V** the all-repos batch menu (fetch/pull/push
everything), **f** a concurrent fetch-all, **a** the Actions menu (theme, refresh), **r** refresh
(rescan). It is deliberately not a full git client — `pull` is fast-forward-only and anything
needing a decision fails cleanly and sends you to a terminal.

### `repos` subcommand

Run a shell command in every git repo nested under a directory (the non-interactive counterpart
to the all-repos screen). The command is joined and run via `sh -c`, so pipes and `&&` work when
quoted as one argument:

```
repoview repos                          # list every nested repo (base-relative paths)
repoview repos -- git status -s         # run in each; header + output only when non-empty
repoview repos --raw -- git fetch       # live-stream each repo's output under its header
repoview repos --dirty -- git pull      # restrict to repos with uncommitted changes
repoview repos -C /path --depth 3 -- pwd
```

Flags: `-C/--dir` (scan root, default cwd), `--raw` (stream vs the default capture), `--dirty`,
`--depth` (default 1). More subcommands (`install`, `update`) are planned.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/brohd11/repoview/main/install.sh | sh
```

Installs into `~/.local/bin`, and offers to add that to your `PATH` if it isn't already.
Prefer to read before you pipe to a shell? Same thing in two steps:

```bash
curl -fsSL -o install.sh https://raw.githubusercontent.com/brohd11/repoview/main/install.sh
less install.sh && sh install.sh
```

Overrides: `BIN_DIR=/usr/local/bin` to install elsewhere, `VERSION=v0.1.1` to pin a release,
`--modify-path` to update your shell rc file without prompting (for unattended setup scripts), or
`--no-modify-path` to leave rc files alone.

Covers macOS (arm64/amd64) and Linux (amd64/arm64). On **Windows**, grab the `.zip` from the
[Releases](https://github.com/brohd11/repoview/releases) page.

With a Go toolchain:

```bash
go install github.com/brohd11/repoview@latest
```

<sub>macOS note: a binary downloaded **in a browser** gets quarantined by Gatekeeper — clear it
with `xattr -dr com.apple.quarantine path/to/repoview`. This doesn't apply to the installer
above; the attribute is set by browsers, not by `curl`.</sub>

Built on [bubblestack](https://github.com/brohd11/bubblestack) (TUI framework) and
[gitstack](https://github.com/brohd11/gitstack) (git engine + screens); it's the manifest-free
sibling of the Godot addon manager [gdaddon](https://github.com/brohd11/gdaddon).

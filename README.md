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

Set `REPOVIEW_DEPTH` to the depth you always want and both the TUI and `repos` start there
instead of 1 — `export REPOVIEW_DEPTH=2` in your shell profile, and `repoview` scans two
levels without a number every time. Anything typed still wins: a positional depth beats
`--depth`, which beats the variable. A malformed or negative value is refused rather than
quietly ignored, and a blank one (`REPOVIEW_DEPTH= repoview`) drops it for a single run.

Keys: **enter** (or **v**) opens the highlighted repo's git menu (status/fetch/pull/push/commit),
**t** opens a terminal at that repo's directory, **V** the all-repos batch menu (fetch/pull/push
everything), **f** a concurrent fetch-all, **a** the Actions menu (theme, self-update, refresh), **r** refresh
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
`--depth` (default `$REPOVIEW_DEPTH`, else 1).

## Install

Unix:
```bash
curl -fsSL https://raw.githubusercontent.com/brohd11/repoview/main/install.sh | sh
```

Windows:

```powershell
irm https://raw.githubusercontent.com/brohd11/repoview/main/install.ps1 | iex
```

To update:
```
repoview update
```

More install details (location, flags, etc): [shared install reference](https://github.com/brohd11/goutil/blob/main/docs/install.md).

<sub>macOS note: a binary downloaded **in a browser** gets quarantined by Gatekeeper — clear it
with `xattr -dr com.apple.quarantine path/to/binary`. This doesn't apply to the installer
above; the attribute is set by browsers, not by `curl`.</sub>
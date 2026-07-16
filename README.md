# repoview

A manifest-free git repo-status viewer for a folder full of projects. Point it at a directory
and it scans for every nested git checkout (fresh each run) and shows each one's status — branch,
uncommitted changes, ahead/behind — in a single TUI list, with fetch/pull/push/commit a keypress
away.

```
repoview [dir]      # dir defaults to the current directory
  -depth N          # max directory depth to scan (default 5)
```

Keys: **enter** opens the highlighted repo's git menu (status/fetch/pull/push/commit), **V** the
all-repos batch menu (fetch/pull/push everything), **f** a concurrent fetch-all, **a** the Actions
menu (theme, refresh), **r** refresh (rescan). It is deliberately not a full git client — `pull`
is fast-forward-only and anything needing a decision fails cleanly and sends you to a terminal.

## Install

```bash
go install github.com/brohd11/repoview@latest
```

Built on [bubblestack](https://github.com/brohd11/bubblestack) (TUI framework) and
[gitstack](https://github.com/brohd11/gitstack) (git engine + screens); it's the manifest-free
sibling of the Godot addon manager [gdaddon](https://github.com/brohd11/gdaddon).

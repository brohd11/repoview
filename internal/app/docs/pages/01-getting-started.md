# Getting started

repoview scans a directory for git checkouts and shows each one's status in one list.

## The idea

A working directory tends to fill up with git repos — a handful of projects side by side,
submodules nested a few levels down, a scratch clone you forgot about. `git status` tells
you about one at a time.

repoview scans a directory, finds every git checkout under it, and shows them together:
each repo's branch, and whether it has uncommitted changes or has drifted from its
upstream. There's no config and no manifest — it re-scans fresh every run.

## Starting up

Run `repoview` in a directory, or give it a path: `repoview <dir>`. By default it looks
one level deep (that directory's immediate children); pass a depth to go deeper:

```
repoview            # current dir, depth 1
repoview 4          # current dir, depth 4
repoview /path 3    # /path, depth 3
```

Positional args are order-free: an all-digits argument is the depth, anything else is the
directory. The header shows the scanned root and how many checkouts it found. When the root
itself is a git checkout, its own status rides the header's `Root:` line.

## The list

One row per repo: its path under the scanned root as the name, its branch as the line
below. A row marks what needs attention:

- `uncommitted changes` — the working tree is dirty
- `ahead 2` — committed locally but not pushed
- `behind origin 3` — someone pushed and you haven't pulled

The ahead/behind counts come straight from git's remote-tracking refs, so they're only as
current as the last fetch — press `f` to fetch every repo and refresh them. repoview never
fetches on its own.

Sort the list with `s`: A→Z, Z→A, then by status (repos needing attention first). Filter
with `/`. `r` re-scans the directory from disk.

## Doing git work

Open a repo with `enter` (or `v`) for its git menu — status, fetch, pull, push, commit —
each streaming git's output to the log pane. `V` does fetch/pull/push across every repo at
once. See the Git menu page for the details.

## Where to next

- Controls — the full key reference
- Git menu — fetch, pull, push, and commit

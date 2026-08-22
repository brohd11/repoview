# Controls

Every key repoview responds to, grouped by what it acts on.

## The repo list

- `enter` / `v` — open the highlighted repo's git menu (status, fetch, pull, push, commit)
- `d` — open the highlighted repo's diff, skipping the menu
- `t` — open a terminal at the repo's directory **in this window**: repoview steps aside,
  your shell takes over, and the list comes back when you `exit`
- `T` — open a terminal at the repo's directory in a **new window**, leaving repoview up
- `ctrl+t` — open the repo's directory in the OS file manager
- `f` — fetch every repo concurrently, refreshing the ahead/behind markers
- `V` — the all-repos git menu: fetch, pull, or push across every repo at once
- `ctrl+v` — the scanned root's own git menu (when the base directory is itself a checkout)
- `s` — cycle the sort order: A→Z, Z→A, by status
- `a` — the Actions menu (theme, docs, update, refresh)
- `r` — re-scan the directory and refresh git state

## Moving around

- `up` / `down` or `k` / `j` — move the cursor
- `g` / `G` — jump to the top / bottom
- `/` — filter the list; `esc` clears it
- `esc` — step back out of any screen
- `q` — quit

## The log pane

The git menus stream their output into a pane below the list:

- `o` — show or hide the pane
- `tab` — focus the pane so you can scroll it
- `w` — wrap its long lines instead of clipping them
- `C` — clear it

## The mouse

- click the header — the scanned root's own git menu (same as `ctrl+v`)
- click a breadcrumb segment — jump back to that screen
- the wheel scrolls the log pane (and focuses it); click the list to hand the keys back
- `ctrl+g` — turn mouse capture off, restoring the terminal's own text selection

## Help

- `?` — expand the help bar to show every key for the current screen

Row keys (`v`, `d`, `t`, `T`, `ctrl+t`) act on the highlighted repo; `V`, `f`, `s`, `a` act
on the list as a whole. `t`/`T`/`ctrl+t` also work from inside a repo's git menu or diff, on
whatever directory that screen is showing. The scheme matches gdaddon, its sibling tool.

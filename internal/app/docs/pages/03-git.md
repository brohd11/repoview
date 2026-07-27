# Git menu

Open a repo with `enter` or `v` for the round-trip: status, fetch, pull, push, and commit,
each streaming git's own output to the log pane.

## Per-repo

- `status` — `git status`, so you can see what's changed before acting
- `fetch` — update the remote-tracking refs, refreshing the ahead/behind counts
- `pull` — `--ff-only`, so a branch that has diverged aborts having changed nothing rather
  than dropping you into a merge conflict inside a TUI
- `push` — send your local commits upstream
- `commit` — asks for a message and what to stage (see below)

It's not a git client and doesn't try to be — an operation that needs a decision from you
refuses instead of guessing. When something fails, git says why in the log, and you go sort
it out in a terminal (`t` opens one at the repo's directory).

## Commit staging

Commit asks what to stage, because git's own default is a trap: `-a` stages changes to
files git already tracks, so a file you *just created* is untracked and would miss the
commit. The confirm screen lists what's going in — and names the new files being left out,
so you can switch to "all, incl. new files" if that's what you meant.

## Across every repo

`V` opens the all-repos git menu: fetch, pull, or push every scanned repo in one go. The
confirm names every repo it will touch, then each runs in turn with its output under its
own header. If one repo refuses — a divergence, say — it's skipped and the rest still run.

`f` on the list is the quick path: a concurrent fetch of every repo with no confirm, just to
refresh the ahead/behind markers. It's the one thing that goes to the network without being
asked for a fuller operation.

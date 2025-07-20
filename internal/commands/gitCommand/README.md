# Git

Portable git operations, like pruning local branches that have been deleted from the remote, or automating sparse checkouts.

## Usage

Run with`--help` to see help menu & args.

## Subcommands

### info

Usage: `syst git info`

Show information about the current Git repository. Assumes the current path is a git repository (and checks before running).

### prune

Usage: `syst git prune [flags]`

Prune local branches that have been deleted from the remote.

Flags:

| Flag                            | Purpose                                           |
| ------------------------------- | ------------------------------------------------- |
| `--confirm`                     | Prompt before deleting each branch                |
| `--dry-run`                     | Show information about the current Git repository |
| `--force`                       | Force delete branches using `git branch -D`       |
| `--main-branch` `[branch-name]` | The name of your main branch (default: `main`)    |

### sparse-clone

Usage: `syst git sparse-clone [flags]`

Clone a git repo with sparse checkout in one step.

A sparse checkout is usually done with the following steps (or something similar):

* `git clone --no-checkout https://remote.com/user/repo.git path/on/localhost`
* `cd path/on/localhost`
* `git sparse-checkout init --cone`
* `git sparse-checkout set path/to/checkout path2/to/checkout ...`
* `git checkout <branch-name>`

Whether this command actually simplifies anything or not, it at least "chains" the commands together to avoid some user error.

Flags:

| Flag                                 | Purpose                                                               |
| ------------------------------------ | --------------------------------------------------------------------- |
| `-b/--checkout-branch [branch-name]` | Branch name to checkout (default: `main`)                             |
| `-p/--checkout-path [path]`          | Paths to sparse-checkout (repeatable, i.e. `-p path/one -p path/two`) |
| `-o/--output-dir [path/on/host]`     | Output directory (defaults to repo name)                              |
| `--protocol [https/ssh]`             | Clone protocol: `ssh` or `https` (default: `ssh`)                     |
| `--provider [provider-name]`         | Git provider (`github`, `gitlab`, `codeberg`) (default: `github`)     |
| `-r/--repository [repo-name]`        | Repository name                                                       |
| `-u/--username [user-or-org-name]`   | Git username or org                                                   |

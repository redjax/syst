# Go System Tool <!-- omit in toc -->

<!-- Repo image -->
<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset=".assets/img/toolbox.png">
    <img src=".assets/img/toolbox.png" height="300">
  </picture>
</p>

<!-- Badges/shields -->
<p align="center">
  <img alt="GitHub Created At" src="https://img.shields.io/github/created-at/redjax/syst">
  <img alt="GitHub Last Commit" src="https://img.shields.io/github/last-commit/redjax/syst">
  <img alt="GitHub Commits this Year" src="https://img.shields.io/github/commit-activity/y/redjax/syst">
  <img alt="Github Repo Size" src="https://img.shields.io/github/repo-size/redjax/syst">
</p>
<!-- Health badges/shields -->
<p align="center">
  <span>🩺 Healthchecks: </span>
  <img alt="Github CodeQL Workflow Status" src="https://img.shields.io/github/actions/workflow/status/redjax/syst/codeql-analysis.yml?branch=main&label=codeQL&labelColor=teal">
  <img alt="GitHub OSV Scan Workflow Status" src="https://img.shields.io/github/actions/workflow/status/redjax/syst/osv-scan.yml?branch=main&label=osvScan&labelColor=maroon">
  <img alt="GitHub Secrets Scan Workflow Status" src="https://img.shields.io/github/actions/workflow/status/redjax/syst/secrets-scan.yml?branch=main&label=secretsScan&lablColor=silver">

</p>

`syst` is a CLI system utility tool. There are multiple subcommands for getting/showing platform info, creating `.zip` backups, and pinging sites/hosts. More functionality will be added over time.

>[!WARNING]
>
> This app is a personal project I'm using to learn Go. There is no warranty or guarantee when using this app, use at your own risk.

## Table of Contents <!-- omit in toc -->

- [Install](#install)
  - [Install script](#install-script)
  - [From Release](#from-release)
  - [Build Locally](#build-locally)
- [Install shell completion](#install-shell-completion)
- [Security Scans](#security-scans)
- [Upgrading](#upgrading)
- [Usage](#usage)
  - [Commands](#commands)
- [Uninstalling](#uninstalling)
- [Reinstalling](#reinstalling)

## Install

### Install script

You can download & install `syst` on Linux & Mac using [the `install-syst.sh` script](./scripts/install-syst.sh).

To download & install in 1 command, do:

```bash
curl -LsSf https://raw.githubusercontent.com/redjax/syst/refs/heads/main/scripts/install-syst.sh | bash -s -- --auto
```

For Windows, use:

```powershell
& ([scriptblock]::Create((irm https://raw.githubusercontent.com/redjax/syst/refs/heads/main/scripts/install-syst.ps1))) -Auto
```

### From Release

Install a release from the [releases page](https://github.com/redjax/syst/releases/latest). You can check the [Verfied Manual Release pipeline](https://github.com/redjax/syst/actions/workflows/create-verified-release.yml) to see the results of a secret & vulnerability scan done before the release.

### Build Locally

Clone this repository and run [one of the build scripts](./scripts/build/).

## Install shell completion

To enable tab-completion for `syst` commands, run one of the following:

- **Bash**: `syst completion bash > ~/.local/share/bash-completion/completions/syst`
- **Zsh**: `syst completion zsh > ~/.local/share/zsh/completions/_syst`
  - Add the completion to your `$fpath` in your `.zshrc`:
    - `fpath=(~/.local/share/zsh/completions $fpath)`
    - Then initialize your completions by adding these lines to `~/.zshrc`:
      - `autoload -U compinit`
      - `compinit`
- **Powershell**: `syst completion powershell > syst.ps1`
  - Open your `$PROFILE` path by running: `ii (Split-Path -Path $PROFILE -Parent)`
  - Create a directory, i.e. `completions/`
  - Move the `syst.ps1` into the `completions/` directory
  - In your `$PROFILE`, source the completions by adding the following:
    - `$PROFILE_PATH = (Split-Path -Path $PROFILE -Parent)`
    - `. "$($PROFILE_PATH)/completions/syst.ps1`

## Security Scans

Each [release](https://github.com/redjax/syst/actions/workflows/create-verified-release.yml) undergoes a scan for secret leaks & Go vulnerabilities (insecure code, malicious dependencies, etc) before a release is created. If the scans detect a vulnerability, the release is cancelled.

This repository also undergoes nightly scans for secret leaks, vulnerabilities, and codeQL. You can see the results in one of the following places:

| Scan | Description |
| ---- | ----------- |
| [codeQL Analysis](./.github/workflows/codeql-analysis.yml) | [codeQL](https://codeql.github.com/docs/codeql-overview/about-codeql/) does ["variant analysis"](https://codeql.github.com/docs/codeql-overview/about-codeql/#about-variant-analysis) on code to detect problems similar to existing vulnerabilities. |
| [OSV scan](./.github/workflows/osv-scan.yml) | [Open Source Vulnerabilities (OSV)](https://osv.dev) is a distributed vulnerabilities database for open source projects. |
| [secrets scan](./.github/workflows/secrets-scan.yml) | Scans repository for strings that look like secrets. Uses the [gitlinks](https://github.com/gitleaks/gitleaks) scanner. |
| [vulnerability scan](./.github/workflows/vulnerability-scan.yml) | Scans Go code for vulnerabilities using the [`govulncheck` Github Action](https://github.com/Templum/govulncheck-action). |

## Upgrading

The CLI includes a `self` subcommand, which allows for running `syst self upgrade` to download a new release. The new version will be downloaded to `syst.new` in the same path as the existing `syst` binary, and on the next execution `syst` will replace the old binary with the new one.

## Usage

Run `syst --help` to print the help menu. For each subcommand, i.e. `syst show`, you can also run `--help` to see scoped parameters for that subcommand.

Each [subcommand](./internal/commands/) has a `README.md` file explaining its purpose/usage.

### Commands

Browse the [commands/ directory](./internal/commands/) to read more about subcommands for this CLI.

## Uninstalling

On Linux, run `sudo rm $(which syst)` to uninstall `syst`.

## Reinstalling

If you have an issue with the `self upgrade` command, you can uninstall & reinstall `syst` with this:

Linux:

```bash
command -v syst >/dev/null 2>&1 && sudo rm "$(command -v syst)" &>/dev/null
curl -LsSf https://raw.githubusercontent.com/redjax/syst/refs/heads/main/scripts/install-syst.sh | bash -s -- --auto
```

You can also [uninstall `syst`](#uninstalling), then reinstall using the one-liner at the top of this page.

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
- [Upgrading](#upgrading)
- [Usage](#usage)
  - [Commands](#commands)

## Install

### Install script

You can download & install `syst` on Linux & Mac using [the `install-syst.sh` script](./scripts/install-syst.sh).

To download & install in 1 command, do:

```bash
curl -LsSf https://raw.githubusercontent.com/redjax/syst/refs/heads/main/scripts/install-syst.sh | bash -s -- --auto
```

### From Release

Install a release from the [releases page](https://github.com/redjax/syst/releases/latest).

### Build Locally

Clone this repository and run [one of the build scripts](./scripts/build/).

## Upgrading

The CLI includes a `self` subcommand, which allows for running `syst self upgrade` to download a new release. The new version will be downloaded to `syst.new` in the same path as the existing `syst` binary, and on the next execution `syst` will replace the old binary with the new one.

## Usage

Run `syst --help` to print the help menu. For each subcommand, i.e. `syst show`, you can also run `--help` to see scoped parameters for that subcommand.

### Commands

Browse the [commands/ directory](./internal/commands/) to read more about subcommands for this CLI.

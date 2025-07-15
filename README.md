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

`syst` is a CLI system utility tool. There are multiple subcommands for getting/showing platform info, creating `.zip` backups, and pinging sites/hosts. More functionality will be added over time.

>[!WARNING]
>
> This app is a personal project I'm using to learn Go. There is no warranty or guarantee when using this app, use at your own risk.

## Table of Contents <!-- omit in toc -->

- [Install](#install)
  - [From Release](#from-release)
  - [Build Locally](#build-locally)
- [Usage](#usage)
  - [Commands](#commands)

## Install

### From Release

Install a release from the [releases page](https://github.com/redjax/syst/releases/latest).

### Build Locally

Clone this repository and run [one of the build scripts](./scripts/build/).

## Usage

Run `syst --help` to print the help menu. For each subcommand, i.e. `syst show`, you can also run `--help` to see scoped parameters for that subcommand.

### Commands

Browse the [commands/ directory](./internal/commands/) to read more about subcommands for this CLI.

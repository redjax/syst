# Version

Print the app's release version. This is built into the app when the `go build` command is run using `-ldflags`. The [build scripts](../../../scripts/build/) and the [Github release Action](../../../.github/workflows/create-release.yml) extract the Git tag, commit, and build date and bake them into the app to be displayed with `--version`.

## Usage

Run `syst --version` to see the CLI's current release

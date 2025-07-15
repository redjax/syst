# ZipBackup

Creates a `.zip` backup of a source directory at a given destination directory.

## Usage

Run with `--help` to see help menu & args.

The command has a `backup` subcommand; to see options, run `syst zipbak --help`.

Args:

| Arg | Purpose |
| --- | ------- |
| `-s/--src` | The source directory to backup |
| `-o/--output-dir` | The output directory where the `.zip` backup will be created/saved |
| `-f/--filename` | The name for the backup `.zip` file |
| `-i/--ignore-paths` | Paths to ignore during the backup; these files will not be copied into the `.zip` archive |
| `-d/--dry-run` | Run the app and print what would happen without taking any actual actions |
| `-c/--cleanup` | When present, do cleanup operations. Meant to be used with `-k` |
| `-k/--keep-backups` | Integer representing the number of backups that should be saved when `--cleanup` runs. Deletes the oldest file(s) first |

## Examples

* `syst zipbak backup -p ~/ -o /opt/backup/home -f home`
  * Create a backup of the user's `$HOME` directory, output to path `/opt/backup/home/YYYY-MM-DD_HH-mm-ss_home.zip`
* `syst zipbak backup -p ~/ -o /opt/backup/home -f home -i .local`
  * Create a backup of the user's `$HOME` directory, output to path `/opt/backup/home/YYYY-MM-DD_HH-mm-ss_home.zip`, skip backup of `~/.local` directory
* `syst zipbak backup -p ~/ -o /opt/backup/home -f home -i .local --cleanup -k 3`
  * Create a backup and run cleanup operations at the end. Retain 3 most recent backups

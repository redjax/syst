# Scanpath

Scan a path for files, optionally matching a filter and/or sorting the results. Prints a table to the CLI with results at the end.

## Usage

Run with `--help` to see help menu & args.

Args:

| Arg | Purpose |
| --- | ------- |
| `-p/--path` | The target path to scan |
| `-l/--limit` | Limit scanned items to reduce scantime |
| `-s/--sort` | Column to sort by (name, size, created, modified, owner, permissions) |
| `-o/--order` | The sort order (asc/desc) |
| `-f/--filter` | A string telling the command how to sort results (e.g. 'size <10MB', 'created >2022-01-01') |

## Examples

* `syst scanpath -p ~/`
  * Scan all files in the home directory
* `syst scanpath -p ~/ -s size -f ">100MB"`
  * Scan all paths in home directory, sort on size and show only files larger than 100MB
* `syst scanpath -p ~/ -s created -f "<2025-07-15" -o asc`
  * Scan all paths in home directory, sort on created date later than July 15th 2025, show oldest items first

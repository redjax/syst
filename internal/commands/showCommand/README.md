# Show

Show information gathered at runtime, i.e. that platform or constants.

## Usage

Run with `--help` to see help menu & args.

Args:

| Arg | Purpose |
| --- | ------- |
| `constants` | Show constants the app dynamically built at runtime |
| `platform` | Show platform info detected by app at runtime |

## Examples

* `syst show constants`
  * Show all constants detected at runtime
* `syst show platform`
  * Show all platform info
* `syst show platform --property hostname --property os --property release --property defaultshell`
  * Show the detected hostname, OS and release, and default shell

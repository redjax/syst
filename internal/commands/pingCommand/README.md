# Ping

Ping utility for pinging websites, hostnames, IP addresses, and FQDNs.

## Usage

Run with `--help` to see help menu & args.

Args:

| Arg          | Purpose                                                                                                                |
| ------------ | ---------------------------------------------------------------------------------------------------------------------- |
| `-c/--count` | The number of pings to send                                                                                            |
| `-s/--sleep` | Count (in seconds) to pause between pings                                                                              |
| `--http`     | Send HTTP HEAD request instead of ICMP ping (for websites)                                                             |
| `--log-file` | When present, a log file will be created at the OS's TEMP path, and the path will be printed at the end of the command |

## Examples

* `syst ping 192.168.1.1 -c 15 -s 5`
  * Ping an IP address 15 times, sleeping for 5 seconds between pings
* `syst ping www.example.com --http`
  * Ping a website indefinitely
* `syst ping hostname.domain -c 0 -s 30`
  * Ping a hostname indefinitely, sleeping for 30 seconds between pings

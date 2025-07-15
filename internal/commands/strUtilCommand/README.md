# Str Utils

String utilities, like capitalization, search/replace, character removal, & substring searching.

## Usage

Run with `--help` to see help menu & args.

Args:

| Arg | Purpose |
| --- | ------- |
| `--capitalize` | Capitalize a string |
| `--upper` | UPPERCASE a string |
| `--lower` | lowercase a string |
| `--title` | Title Case a String |
| `-i/--ignore-case` | Ignore capitalization for search/replace & removal operations |
| `--remove` | Remove occurrences of a character in a string |
| `--replace` | Search a string for a substring and replace it with another string |
| `--search` | Search a string for a given substring |


## Examples

* `syst strutil "This is a string" --title`
  * TitleCase the String
* `syst strutil "This is a string" --lower`
  * lowercase the string
* `syst strutil "This is a string and this part will be removed" --remove " and this part will be removed"`
  * Remove a substring from an input string
* `syst strutil "This is a string ^, the carat symbol will be replaced with *" --replace "^/*"`
  * Replace all instances of `^` in the string with `*`
* `syst strutil "This is a string, and this is the part to search for" --search "search for"`
  * Search a string for a substring
* `cat path/to/somelogfile.log | syst strutil --search "this is the string to find" -i`
  * Read the results of the `cat` operation and do case-insensitive substring search on the contents

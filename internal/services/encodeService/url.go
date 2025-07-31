package encodeservice

import "net/url"

func EncodeStringUrl(input string) string {
	return url.QueryEscape(input)
}

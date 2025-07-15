// below ping.go (add to bottom or make a new ping_errors.go)
package pingService

import "errors"

var (
	ErrEmptyTarget = errors.New("target must not be empty")
)

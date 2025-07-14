//go:build !linux && !windows && !darwin
// +build !linux,!windows,!darwin

package constants

func platformConstants() PlatformConstants {
	return PlatformConstants{}
}

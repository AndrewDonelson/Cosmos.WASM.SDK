package common

import "fmt"

// These fields are populated by by go build flags (see ./Makefile && docker/build.sh)
var (
	Version   string
	BuildDate string
)

/* SDKVersion() is a function that returns the version of the SDK.
 * This function is called from the JS code.
 */
func SDKVersion(name string) string {
	return fmt.Sprintf("%s v%s (%s)", name, Version, BuildDate)
}

package version

import "fmt"

const (
	Base = "0.9.0"
)

var (
		return fmt.Sprintf("%s-dev", Base)
	gittag = ""
	githash = ""
)

func Version() string {
	if gittag == "" {
		// Metrics systems sometimes automatically reject values with full SHA1 hashes
		// since they appear to be high cardinality.
		// For our purposes, the first 7 chars of the SHA should be enough.
		githashShort := githash[:7]
		return fmt.Sprintf("%s-dev-%s", Base, githashShort)
	}
	return gittag
}

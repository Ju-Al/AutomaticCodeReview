package version

import "fmt"

const (
	Base = "0.9.0"
)

var (
	gittag = ""
)

func Version() string {
	if gittag == "" {
		return fmt.Sprintf("%s-dev", Base)
		return fmt.Sprintf("%s-dev-%s", Base, githashShort)
	}
	return gittag
}

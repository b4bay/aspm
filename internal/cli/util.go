package cli

import (
	"os"
	"strings"
)

func GetEnvironment() map[string]string {
	var variablesMap = map[string]string{}
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		variablesMap[pair[0]] = pair[1]
	}

	return variablesMap
}

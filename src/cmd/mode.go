package cmd

import (
	"fmt"
	"strings"
)

func Mode(args ...string) string {
	nickname := args[0]
	flags := strings.Join(args[1:], " ")

	return fmt.Sprintf("MODE %v %v", nickname, flags)
}

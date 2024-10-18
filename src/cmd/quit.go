package cmd

import (
	"fmt"
	"strings"
)

func Quit(args ...string) string {
	return fmt.Sprintf("QUIT %v", strings.Join(args, " "))
}

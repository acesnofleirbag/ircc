package cmd

import (
	"fmt"
	"strings"
)

func Squit(args ...string) string {
	server := args[0]
	msg := strings.Join(args[1:], " ")

	return fmt.Sprintf("SQUIT %v %v", server, msg)
}

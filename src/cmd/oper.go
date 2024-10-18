package cmd

import (
	"fmt"
)

func Oper(args ...string) string {
	name := args[0]
	passwd := args[1]

	return fmt.Sprintf("OPER %v %v", name, passwd)
}

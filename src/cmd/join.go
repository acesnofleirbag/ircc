package cmd

import "fmt"

func Join(args ...string) string {
	server := args[0]

	return fmt.Sprintf("JOIN %v", server)
}

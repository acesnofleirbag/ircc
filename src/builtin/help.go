package builtin

func Help(args ...string) []string {
	return []string{
		"== HELP ==",
		"-",
		"Supported Commands:",
		"-",
		"list-flags: List all flags to MODE command",
	}
}

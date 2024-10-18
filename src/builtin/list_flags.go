package builtin

func ListFlags(args ...string) []string {
	return []string{
		"== LIST FLAGS ==",
		"-",
		"[ a ] user is flagged as away",
		"[ i ] marks a users as invisible",
		"[ w ] user receives wallops",
		"[ r ] restricted user connection",
		"[ o ] operator flag",
		"[ O ] local operator flag",
		"[ s ] marks a user for receipt of server notices (obsolete)",
	}
}

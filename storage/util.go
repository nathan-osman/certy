package storage

func ifProvided(v string) []string {
	if v == "" {
		return []string{}
	}
	return []string{v}
}

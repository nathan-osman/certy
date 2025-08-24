package storage

func (s *Storage) ifProvided(v string) []string {
	if v == "" {
		return []string{}
	}
	return []string{v}
}

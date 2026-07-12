package app

import "strings"

// containsCI reports whether target is present in list, case-insensitively.
func containsCI(list []string, target string) bool {
	for _, item := range list {
		if strings.EqualFold(item, target) {
			return true
		}
	}
	return false
}

func joinOr(list []string) string {
	return strings.Join(list, " or ")
}

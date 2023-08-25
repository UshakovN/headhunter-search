package utils

func OneOf(f func(s string) bool, s ...string) bool {
	for _, s := range s {
		if f(s) {
			return true
		}
	}
	return false
}

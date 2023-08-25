package utils

import "strings"

func TrimPrefixIfExist(s string, prefix string) (string, bool) {
	st := strings.TrimPrefix(s, prefix)
	ok := strings.HasPrefix(s, prefix)
	return st, ok
}

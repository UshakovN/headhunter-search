package str

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var regexWord = regexp.MustCompile(`[^а-яА-Яa-zA-Z]`)

func OneOf(f func(s string) bool, s ...string) bool {
	for _, s := range s {
		if f(s) {
			return true
		}
	}
	return false
}

func MustCast[T any](s string) T {
	var (
		iface any
		err   error
	)
	switch typ := any(*new(T)).(type) {
	case string:
		iface = any(s).(T)
	case int:
		iface, err = strconv.Atoi(s)
	case int64:
		iface, err = strconv.ParseInt(s, 10, 64)
	default:
		err = fmt.Errorf("unsupported type %v", typ)
	}
	if err != nil {
		panic(err)
	}
	return iface.(T)
}

func BuildSentenceTags(s string) []string {
	s = sanitizeSentence(s)

	words := strings.Split(s, " ")
	tags := make([]string, 0, len(words))

	tag := strings.Join(words, "_")
	tags = append(tags, fmt.Sprintf("#%s", strings.ToLower(tag)))

	for _, word := range words {
		word = strings.TrimSpace(word)
		word = strings.ToLower(word)

		tag := fmt.Sprintf("#%s", word)
		tags = append(tags, tag)
	}
	return tags
}

func sanitizeSentence(s string) string {
	s = regexWord.ReplaceAllLiteralString(s, " ")
	s = regexPrimary.ReplaceAllLiteralString(s, " ")
	s = regexSecondary.ReplaceAllLiteralString(s, " ")
	s = strings.TrimSpace(s)
	return s
}

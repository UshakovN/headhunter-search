package str

import (
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var (
	strictSanitizer = bluemonday.StrictPolicy()
	regexPrimary    = regexp.MustCompile(`[\t\r\n]| {2,}`)
	regexSecondary  = regexp.MustCompile(`\s{2,}`)
)

func Sanitize(s string) string {
	s = strictSanitizer.Sanitize(s)
	s = regexPrimary.ReplaceAllLiteralString(s, " ")
	s = regexSecondary.ReplaceAllLiteralString(s, " ")
	s = strings.TrimSpace(s)
	return s
}

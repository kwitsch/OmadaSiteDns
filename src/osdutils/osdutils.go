package osdutils

import "regexp"

var (
	validName  = regexp.MustCompile("^[a-z,0-9][a-z,0-9,-]*[a-z,0-9]")
	validSub   = regexp.MustCompile("[a-z,0-9,-]")
	remInvalid = regexp.MustCompile("[^a-z,0-9,-]")
)

func ValidDnsStr(input string) bool {

	if len(input) < 253 {
		return validName.MatchString(input)
	}
	return false
}

func ValidSubstitute(input string) bool {
	return validSub.MatchString(input)
}

func RemoveInvalidCharacters(input string) string {
	return remInvalid.ReplaceAllString(input, "")
}

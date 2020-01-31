package utils

import "strings"

func Chomp(text string) string {
	return strings.Replace(text, "\n", "", -1)
}

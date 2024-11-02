package ui

import (
	"unicode"
	"unicode/utf8"
)

type accept func(textToCheck string, lastChar rune) bool

func isNumber() accept {
	return func(textToCheck string, lastChar rune) bool {
		return unicode.IsDigit(lastChar)
	}
}

func length(l int) accept {
	return func(textToCheck string, lastChar rune) bool {
		return utf8.RuneCountInString(textToCheck) <= l
	}
}

func summCheck(accepts ...accept) accept {
	return func(textToCheck string, lastChar rune) bool {
		for _, f := range accepts {
			if !f(textToCheck, lastChar) {
				return false
			}
		}
		return true
	}
}

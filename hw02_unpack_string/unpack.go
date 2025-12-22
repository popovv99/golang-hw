package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(str string) (string, error) {
	var builder strings.Builder
	var prev rune
	escapeMode := false
	const backSlash = '\\'
	for _, current := range str {
		if current == backSlash && !escapeMode {
			escapeMode = true
			write(&builder, prev)
			continue
		}
		num, err := strconv.Atoi(string(current))
		if escapeMode {
			if err != nil && current != backSlash {
				return "", ErrInvalidString
			}
			prev = current
			escapeMode = false
			continue
		}
		if err == nil {
			if prev == 0 {
				return "", ErrInvalidString
			}
			builder.WriteString(strings.Repeat(string(prev), num))
			prev = 0
		} else {
			write(&builder, prev)
			prev = current
		}
	}
	if escapeMode {
		return "", ErrInvalidString
	}
	write(&builder, prev)
	return builder.String(), nil
}

func write(builder *strings.Builder, r rune) {
	if r != 0 {
		builder.WriteString(string(r))
	}
}

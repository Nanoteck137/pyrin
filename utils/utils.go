package utils

import (
	"strings"
)

type NameMapping func(name string) string
type ReplacementFunc func(name string) string

func ReplacePathArgs(path string, nameMapping NameMapping, replacementFunc ReplacementFunc) (string, []string) {
	var args []string
	parts := strings.Split(path, "/")

	for i, p := range parts {
		if len(p) == 0 {
			continue
		}

		if p[0] == ':' {
			name := p[1:]
			if nameMapping != nil {
				name = nameMapping(name)
			}
			
			args = append(args, name)

			parts[i] = replacementFunc(name)
		}
	}

	newPath := strings.Join(parts, "/")

	return newPath, args
}

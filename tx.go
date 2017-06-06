package main

import (
	"fmt"
	"regexp"
)

var (
	withAddr   *regexp.Regexp = regexp.MustCompile("(.*)\\s+(#.*)")
	withDigits *regexp.Regexp = regexp.MustCompile("(.*)\\s+([0-9]+(-?[0-9]+)+\\s+.*)")
	cities     [10]string     = [10]string{
		"MONTREAL", "MAGOG", "VANCOUVER", "TORONTO", "BOLTON", "CALGARY",
		"MISSISSAUGA", "CANMORE", "OUTREMONT", "CHAMBLY",
	}
)

type Match struct {
	Name string
	Desc string
}

func extractName(n string) *Match {
	matches := withAddr.FindStringSubmatch(n)
	if len(matches) > 2 {
		return &Match{matches[1], matches[2]}
	}

	matches = withDigits.FindStringSubmatch(n)
	if len(matches) > 2 {
		return &Match{matches[1], matches[2]}
	}

	for _, c := range cities {
		withCity := regexp.MustCompile(fmt.Sprintf("(.*)\\s+(%s.*)", c))
		matches := withCity.FindStringSubmatch(n)
		if len(matches) > 2 {
			return &Match{matches[1], matches[2]}
		}
	}

	return nil
}

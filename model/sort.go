package model

import (
	"strings"
)

type Sorts string

func (s Sorts) Order() string {
	split := strings.Split(string(s), ",")
	to := make([]string, len(split))
	for i := range split {
		to[i] = strings.Trim(split[i], `""`)
		to[i] = strings.TrimSpace(to[i])
		if len(to[i]) > 1 {
			if to[i][:1] == "-" {
				to[i] = to[i][1:] + " desc"
			}
		}
	}
	return strings.Join(to, ",")
}

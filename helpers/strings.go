package helpers

import "strings"

func ContainsSub(data string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(candidate, data) {
			return true
		}
	}

	return false
}

func Contains(path string, excludes ...string) bool {
	for _, exclude := range excludes {
		if exclude == path {
			return true
		}
	}

	return false
}

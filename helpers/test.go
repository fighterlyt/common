package helpers

import "flag"

func IsTest() bool {
	return flag.Lookup("test.v") != nil
}

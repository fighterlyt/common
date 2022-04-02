package parser

import (
	"os"
	"testing"

	"gitlab.com/nova_dubai/usdtpay/config"
)

var (
	resource *config.Resource
)

func TestMain(m *testing.M) {
	conf, err := config.LoadConfig("../../../config/conf", 0)
	if err != nil {
		panic(err)
	}

	resource, err = config.LoadResource(conf)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

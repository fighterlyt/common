package summary

import (
	"testing"

	"github.com/fighterlyt/log"
	"gorm.io/gorm"
)

var (
	logger log.Logger
	err    error
	db     *gorm.DB
)

func TestMain(m *testing.M) {
}

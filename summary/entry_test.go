package summary

import (
	"github.com/fighterlyt/gormlogger"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"os"
	"testing"
	"time"

	"github.com/fighterlyt/log"
	"gorm.io/gorm"
)

var (
	logger log.Logger
	err    error
	db     *gorm.DB
)

func TestMain(m *testing.M) {
	if logger, err = log.NewEasyLogger(true, false, ``, `test`); err != nil {
		panic(`构建日志器` + err.Error())
	}

	dsn := "root:dubaihell@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"

	mysqlLogger := gormlogger.NewLogger(logger.Derive(`mysql`).SetLevel(zapcore.DebugLevel).AddCallerSkip(1), time.Second, nil)

	if db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: mysqlLogger,
	}); err != nil {
		panic(`构建数据库` + err.Error())
	}

	os.Exit(m.Run())
}

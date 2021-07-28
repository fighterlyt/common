package message

import (
	"os"
	"testing"
	"time"

	"github.com/fighterlyt/gormlogger"
	"github.com/fighterlyt/log"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db          *gorm.DB
	testLogger  log.Logger
	err         error
	testService *service
)

func TestMain(m *testing.M) {
	if testLogger, err = log.NewEasyLogger(true, false, ``, `test`); err != nil {
		panic(`构建日志器` + err.Error())
	}

	dsn := "root:dubaihell@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"

	mysqlLogger := gormlogger.NewLogger(testLogger.Derive(`mysql`).SetLevel(zapcore.DebugLevel).AddCallerSkip(1), time.Second, nil)

	if db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: mysqlLogger.LogMode(logger.Info),
	}); err != nil {
		panic(`构建数据库` + err.Error())
	}

	os.Exit(m.Run())
}

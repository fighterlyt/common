package summaryextend

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/fighterlyt/gormlogger"
	"gitlab.com/nova_dubai/common/helpers"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"

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

	logger.Debug(`debug`)

	dsn := "root:dubaihell@tcp(127.0.0.1:3306)/first?charset=utf8mb4&parseTime=True&loc=Local"

	targetLogger := logger.Derive(`mysql`)
	targetLogger.SetLevel(zapcore.InfoLevel).AddCallerSkip(1)

	mysqlLogger := gormlogger.NewLogger(targetLogger, time.Second, nil)

	mysqlLogger.Info(context.Background(), `a`)

	if db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: mysqlLogger,
	}); err != nil {
		panic(`构建数据库` + err.Error())
	}

	helpers.SetTimeZone(helpers.GetBeiJin())
	os.Exit(m.Run())
}

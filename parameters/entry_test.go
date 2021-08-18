package parameters

import (
	"gitlab.com/nova_dubai/common/helpers"

	"os"
	"testing"
	"time"

	"github.com/fighterlyt/gormlogger"
	"github.com/fighterlyt/log"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	testService *service
	db          *gorm.DB
	err         error
	testLogger  log.Logger
	redisClient *redis.Client
)

var (
	loggerLevel = zapcore.DebugLevel

	cfg = &log.Config{
		Service:    "测试",
		Level:      loggerLevel,
		FilePath:   "",
		TimeZone:   "",
		TimeLayout: "",
		Debug:      true,
		JSON:       false,
	}
	redisOption = &redis.Options{
		Addr:     "127.0.0.1:9736",
		DB:       2,
		Password: `dubaihell`,
	}
)

func TestMain(m *testing.M) {
	if testLogger, err = cfg.Build(); err != nil {
		panic(`构建日志器` + err.Error())
	}

	dsn := "root:dubaihell@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.NewLogger(testLogger.Derive(`mysql`).SetLevel(zapcore.DebugLevel).AddCallerSkip(3), time.Second, map[string]zapcore.Level{
			`mysql`: loggerLevel,
		}),
	})
	if err != nil {
		panic(`gorm` + err.Error())
	}

	redisClient = redis.NewClient(redisOption)

	redisClient.AddHook(helpers.NewRedisLogger(testLogger))

	Init(`data/parameters.json`, nil)

	if testService, err = NewService(db, redisClient, testLogger, gin.Default()); err != nil {
		panic(`NewService ` + err.Error())
	}

	os.Exit(m.Run())
}

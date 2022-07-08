package model

import (
	"context"
	"fmt"
	"github.com/vmihailenco/msgpack"
	"os"
	"testing"
	"time"

	"github.com/fighterlyt/gormlogger"
	"github.com/fighterlyt/log"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	logger      log.Logger
	err         error
	db          *gorm.DB
	redisClient *redis.Client

	testFromHost = "https://dubai-real.oss-accelerate-overseas.aliyuncs.com"
	testToHost   = "https://d.khols8.com/"

	testOssFilePath = OssFilePath("https://dubai-real.oss-accelerate-overseas.aliyuncs.com/first/icon/contract.png")
	resultPath      = "https://d.khols8.com//first/icon/contract.png"
)

type testOssFile struct {
	ID   int64       `gorm:"column:id;primaryKey;comment:id" json:"id"`
	Link OssFilePath `gorm:"column:link;type:varchar(256);comment:地址" json:"link"`
	Name string      `gorm:"column:name;type:varchar(128);comment:key" json:"name"`
}

// MarshalBinary use msgpack
func (s *testOssFile) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(s)
}

// UnmarshalBinary use msgpack
func (s *testOssFile) UnmarshalBinary(data []byte) error {
	return msgpack.Unmarshal(data, s)
}

func newTestOssFile(id int64, link OssFilePath, name string) *testOssFile {
	return &testOssFile{ID: id, Link: link, Name: name}
}

func (testOssFile) TableName() string {
	return `test_oss_file`
}

func TestMain(m *testing.M) {
	if logger, err = log.NewEasyLogger(true, false, ``, `test`); err != nil {
		panic(`构建日志器` + err.Error())
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:9736",
		DB:       2,
		Password: `dubaihell`,
	})

	dsn := "root:dubaihell@tcp(127.0.0.1:6033)/test?charset=utf8mb4&parseTime=True&loc=Local"

	mysqlLogger := gormlogger.NewLogger(logger.Derive(`mysql`).SetLevel(zapcore.DebugLevel).AddCallerSkip(1), time.Second, nil)

	if db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: mysqlLogger,
	}); err != nil {
		panic(`构建数据库` + err.Error())
	}

	if err = db.AutoMigrate(&testOssFile{}); err != nil {
		panic(`自动生成数据库表失败` + err.Error())
	}

	os.Exit(m.Run())
}

func TestOssFileMysqlInsert(t *testing.T) {
	record := newTestOssFile(1, testOssFilePath, "测试")

	require.NoError(t, db.Save(record).Error)
}

func TestOssFileMysqlFind(t *testing.T) {
	TestOssFileMysqlInsert(t)

	var record = &testOssFile{}

	err = db.Model(&testOssFile{}).Where("id=?", 1).First(record).Error
	require.NoError(t, err)

	if string(record.Link) != string(testOssFilePath) {
		t.Error("文件路径错误")
	}

	fmt.Println("oss路径:", record.Link)
}

func TestOssFileRedisSet(t *testing.T) {
	record := newTestOssFile(1, testOssFilePath, "测试")

	err = redisClient.Set(context.Background(), "oss_file_test", record, time.Minute).Err()
	require.NoError(t, err)
}

func TestOssFileRedisGet(t *testing.T) {
	TestOssFileRedisSet(t)

	record := new(testOssFile)

	err = redisClient.Get(context.Background(), "oss_file_test").Scan(record)
	require.NoError(t, err)

	if string(record.Link) != string(testOssFilePath) {
		t.Error("文件路径错误")
	}

	fmt.Println("oss路径:", record.Link)
}

func TestOssFilePath_MarshalJSON_notSet(t *testing.T) {
	_, err := testOssFilePath.MarshalText()
	require.Error(t, err)
}

func TestOssFilePath_MarshalJSON_OnlyFrom(t *testing.T) {
	SetOssFromHost(testFromHost)

	_, err := testOssFilePath.MarshalText()
	require.Error(t, err)
}

func TestOssFilePath_MarshalJSON_OnlyTo(t *testing.T) {
	SetOssToHost(testToHost)

	_, err := testOssFilePath.MarshalText()
	require.Error(t, err)
}

func TestOssFilePath_MarshalJSON_setAll(t *testing.T) {
	SetOssFromHost(testFromHost)
	SetOssToHost(testToHost)

	result, err := testOssFilePath.MarshalText()
	require.NoError(t, err)

	if string(result) != resultPath {
		panic("替换失败")
	}

	fmt.Println(string(result))
}

func TestOssFilePath_MarshalJSON_setAllEmpty(t *testing.T) {
	SetOssFromHost("")
	SetOssToHost("")

	result, err := testOssFilePath.MarshalText()
	require.NoError(t, err)

	if string(result) != string(testOssFilePath) {
		panic("替换失败")
	}

	fmt.Println(string(result))
}

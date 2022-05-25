package free_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/fighterlyt/gotron-sdk/pkg/client"
	"github.com/fighterlyt/log"
	"github.com/fighterlyt/redislock"
	"github.com/go-redis/redis/v8"
	"gitlab.com/nova_dubai/common/cryptocurrency/tron/free"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	mysqlAddress = `root:dubaihell@tcp(127.0.0.1:3306)/freeze?charset=utf8mb4&parseTime=True&loc=Local`
	tronAddress  = `35.181.32.79:50051`
)

var (
	logger     log.Logger
	err        error
	db         *gorm.DB
	g          *client.GrpcClient
	locker     redislock.Locker
	service    free.Service
	from       = ``
	privateKey = ``
)

func TestMain(m *testing.M) {
	if logger, err = log.NewEasyLogger(true, false, ``, `质押`); err != nil {
		panic(err.Error())
	}

	if db, err = gorm.Open(mysql.Open(mysqlAddress), &gorm.Config{}); err != nil {
		panic(`连接mysql错误` + err.Error())
	}

	g = client.NewGrpcClient(tronAddress)

	if err = g.Start(grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		panic(err.Error())
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:9736",
		Password: "dubaihell",
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()
	if err = redisClient.Ping(ctx).Err(); err != nil {
		panic(`redis` + err.Error())
	}

	locker = redislock.NewLocker(redisClient)

	os.Exit(m.Run())
}

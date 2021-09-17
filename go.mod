module gitlab.com/nova_dubai/common

go 1.16

require (
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/boombuler/barcode v1.0.1
	github.com/bwmarrin/snowflake v0.3.0
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/dgoogauth v0.0.0-20190221195224-5a805980a5f3
	github.com/ethereum/go-ethereum v1.10.7
	github.com/fighterlyt/gormlogger v0.0.0-20210729161641-ae2b6a621523
	github.com/fighterlyt/gotron-sdk v0.0.0-20210726202906-8b77a73e46fb
	github.com/fighterlyt/log v0.0.0-20210607120019-54cae88916e3
	github.com/fighterlyt/redislock v0.0.0-20210520112328-c517c4f54b7f
	github.com/gin-gonic/gin v1.7.4
	github.com/go-redis/redis/v8 v8.11.3
	github.com/go-redsync/redsync/v4 v4.4.1
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/shirou/gopsutil v3.21.8+incompatible
	github.com/shopspring/decimal v1.2.0
	github.com/stretchr/testify v1.7.0
	gitlab.com/nova_dubai/cache v0.0.0-20210824010034-1c70f17b8fe4
	go.uber.org/atomic v1.9.0
	go.uber.org/multierr v1.7.0
	go.uber.org/zap v1.19.1
	gopkg.in/tucnak/telebot.v2 v2.4.0
	gorm.io/driver/mysql v1.1.2
	gorm.io/gorm v1.21.15
)

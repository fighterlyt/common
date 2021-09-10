package helpers

import (
	"github.com/fighterlyt/log"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// BaseResource 基础服务
type BaseResource struct {
	DB          *gorm.DB      // 数据库
	IRoutes     gin.IRoutes   // gin
	RedisClient *redis.Client // redis 客户端
	Logger      log.Logger    // 日志器
	TestIRoutes gin.IRoutes   // 用于测试或者模拟的gin
}

/*NewBaseResource 新建基础服务
参数:
*	db           	*gorm.DB     	数据库
*	iRoutes      	gin.IRoutes  	gin
*   TestIRoutes     gin.IRoutes     测试gin
*	redisClient  	*redis.Client   redis客户端
*	logger       	log.Logger      日志器
返回值:
*	*BaseResource	*BaseResource	基础服务
*/
func NewBaseResource(db *gorm.DB, iRoutes, testIRoutes gin.IRoutes, redisClient *redis.Client, logger log.Logger) *BaseResource {
	return &BaseResource{
		DB:          db,
		IRoutes:     iRoutes,
		RedisClient: redisClient,
		Logger:      logger,
		TestIRoutes: testIRoutes,
	}
}

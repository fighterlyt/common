package parameters

import (
	"encoding/json"
	"fmt"
	"gitlab.com/nova_dubai/common/twofactor"
	"os"
	"time"

	"github.com/fighterlyt/log"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"gitlab.com/nova_dubai/common/helpers"
	"gitlab.com/nova_dubai/common/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	key  = `parameters`
	name = `参数管理`
)

type service struct {
	db                *gorm.DB         // 数据库
	client            *redis.Client    // redis
	parameter         ParameterService // 核心参数服务
	history           HistoryService   // 变更历史服务
	logger            log.Logger       // 日志器
	router            gin.IRouter      // http router
	shutdown          model.Shutdown   // 关闭
	auth              twofactor.Auth   // 验证器
	needTwoFactorKeys []string         // 需要二次验证的key
}

func (s *service) Close() {
	s.shutdown.Close()
}

func (s *service) Add(count int64) {
	s.shutdown.Add(count)
}

func (s *service) IsClosed() bool {
	return s.shutdown.IsClosed()
}

func (s *service) Key() string {
	return key
}

func (s *service) Name() string {
	return name
}

/*NewService 新建服务
参数:
*	db           	*gorm.DB     	数据库
*	client       	*redis.Client	redis
*	logger       	log.Logger      日志器
*	iRouter      	gin.IRouter  	HTTP句柄
返回值:
*	targetService	*service     	返回值1
*	err          	error        	返回值2
*/
func NewService(db *gorm.DB, client *redis.Client, logger log.Logger, iRouter gin.IRouter) (targetService *service, err error) {
	if dataPath == `` {
		return nil, fmt.Errorf(`未初始化，必须调用初始化方法Init()`)
	}

	history := newHistoryService(db, logger.Derive(`变更管理器`))
	parameter := newParameterService(client, db, logger.Derive(`数据管理器`), time.Minute)
	targetService = &service{
		db:        db,
		client:    client,
		history:   history,
		parameter: parameter,
		logger:    logger,
		router:    iRouter,
		shutdown:  model.NewShutdown(),
	}

	moduleLogger = logger

	if err := targetService.init(); err != nil {
		return nil, errors.Wrap(err, `初始化失败`)
	}

	return targetService, nil
}

func (s *service) SetTwoFactorAuth(needTwoFactorKeys []string, auth twofactor.Auth) {
	s.needTwoFactorKeys = needTwoFactorKeys
	s.auth = auth
}

func (s *service) GetParameters(keys ...string) (parameters map[string]*Parameter, err error) {
	return s.parameter.GetParameters(keys...)
}

/*Modify 修改参数
参数:
*	keyValue	map[string]string	key->value map
*	userID  	int64               用户ID
返回值:
*	error   	error            	错误
*/
func (s *service) Modify(keyValue map[string]string, userID int64) error {
	for key, value := range keyValue {
		if err := s.parameter.Modify(key, value); err != nil {
			return errors.Wrap(err, `修改数据`)
		}

		if err := s.history.Save(key, value, userID); err != nil {
			s.logger.Warn(`保存变更记录失败`, zap.String(`错误`, err.Error()))
		}
	}

	return nil
}

/*AddParameters 添加参数
参数:
*	parameters	...*Parameter	参数
返回值:
*	error     	error           错误
*/
func (s *service) AddParameters(parameters ...*Parameter) error {
	for _, parameter := range parameters {
		if err := s.parameter.Save(parameter); err != nil {
			return errors.Wrapf(err, `保存参数[%s]失败`, parameter.Key)
		}

		if err := s.history.Save(parameter.Key, parameter.Value, -1); err != nil {
			return errors.Wrapf(err, `保存[%s]变更记录失败`, parameter.Key)
		}
	}

	return nil
}

/*GetHistory 获取历史
参数:
*	key      	string   	key
*	start    	int      	开始位置，0开始
*	limit    	int      	限量
返回值:
*	histories	[]History	历史变更记录
*	err      	error    	错误
*/
func (s *service) GetHistory(key string, startTime, endTime int64, start, limit int) (allCount int64, histories []History, err error) {
	return s.history.Get(key, startTime, endTime, start, limit)
}

/*Create 创建参数
参数:
*	parameter	*Parameter	参数
返回值:
*	error    	error     	错误
*/
func (s *service) Create(parameter *Parameter) error {
	return s.parameter.Save(parameter)
}

/*init 初始化
参数:
返回值:
*	error	error	错误
*/
func (s *service) init() error {
	if err := s.dbInit(); err != nil {
		return errors.Wrap(err, `数据库初始化失败`)
	}

	if err := s.loadConfig(); err != nil {
		return errors.Wrap(err, `加载配置数据`)
	}

	s.http()

	return nil
}

/*dbInit 数据库初始化
参数:
返回值:
*	error	error	错误
*/
func (s *service) dbInit() error {
	if err := s.db.AutoMigrate(&History{}); err != nil {
		return errors.Wrap(err, `创建变更表`)
	}

	if err := s.db.AutoMigrate(&Parameter{}); err != nil {
		return errors.Wrap(err, `创建记录表`)
	}

	return nil
}

/*loadConfig 加载配置文件
参数:
返回值:
*	error	error	返回值1
*/
func (s *service) loadConfig() error {
	file, err := os.Open(dataPath)
	if err != nil {
		return errors.Wrapf(err, `打开数据文件[%s]`, dataPath)
	}

	defer helpers.IgnoreError(s.logger, `关闭数据文件`, func() error {
		return file.Close()
	})

	var (
		parameters []*Parameter
	)

	if err = json.NewDecoder(file).Decode(&parameters); err != nil {
		return errors.Wrap(err, `JSON读取`)
	}

	for _, parameter := range parameters {
		parameter.UpdateTime = time.Now().Unix()
		if err = parameter.Validate(); err != nil {
			return errors.Wrapf(err, `参数key=[%s]验证失败`, parameter.Key)
		}

		if err = s.Create(parameter); err != nil {
			return errors.Wrapf(err, `参数key=[%s]保存失败`, parameter.Key)
		}
	}

	return nil
}

var (
	dataPath string
)

package free

import (
	"fmt"
	"time"

	"github.com/fighterlyt/gotron-sdk/pkg/client"
	"github.com/fighterlyt/gotron-sdk/pkg/free"
	"github.com/fighterlyt/gotron-sdk/pkg/proto/core"
	"github.com/fighterlyt/log"
	"github.com/fighterlyt/redislock"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.com/nova_dubai/common/helpers"
	"gorm.io/gorm"
)

const (
	// TRXForSingleEnergy 单次转账需要质押的TRX
	TRXForSingleEnergy = 1234
	// TRXToSUN  TRX 和SUN的对应关系
	TRXToSUN = 1000000
)

var (
	lockTimeout = time.Second * 3
)

type service struct {
	db         *gorm.DB
	logger     log.Logger
	from       string
	privateKey string
	tronClient *client.GrpcClient
	locker     redislock.Locker
	Hooks
}

func NewService(db *gorm.DB, logger log.Logger, tronClient *client.GrpcClient, locker redislock.Locker) (s Service, err error) {
	target := &service{
		db:         db,
		logger:     logger,
		tronClient: tronClient,
		locker:     locker,
	}

	if err = target.validate(); err != nil {
		return nil, err
	}

	target.Hooks = NewHooks(logger.Derive(`hooks`))

	if err = target.init(); err != nil {
		return nil, err
	}

	return target, nil
}

func (s service) init() error {
	if err := s.dbInit(); err != nil {
		return errors.Wrap(err, `数据库初始化`)
	}

	return nil
}

func (s service) dbInit() error {
	if err := s.db.AutoMigrate(modelRecord, modelFail); err != nil {
		return errors.Wrap(err, `建表`)
	}

	return nil
}

func (s service) validate() error {
	if s.db == nil {
		return errors.New(`db不能为空`)
	}

	if s.logger == nil {
		return errors.New(`日志器不能为空`)
	}

	if s.tronClient == nil {
		return errors.New(`波场客户端不能为空`)
	}

	if s.locker == nil {
		return errors.New(`redis分布式锁不能为空`)
	}

	return nil
}

/*SetUp 设置
参数:
*	from      	string	执行质押TRX的地址
*	privateKey	string	from对应的私钥
返回值:
*	error     	error 	错误
*/
func (s *service) SetUp(from, privateKey string) error {
	var (
		matched bool
		err     error
	)

	if matched, err = helpers.IsPrivateKeyMatched(from, privateKey); err != nil {
		return errors.Wrap(err, `判断私钥和地址是否符合`)
	}

	if !matched {
		return fmt.Errorf(`私钥地址不符合`)
	}

	s.from, s.privateKey = from, privateKey

	return nil
}

/*Freeze 冻结
参数:
*	to       	string          收益方
*	trxAmount	decimal.Decimal	冻结TRX 金额
返回值:
*	error    	error          	错误
*/
func (s service) Freeze(to string, trxAmount decimal.Decimal) error {
	info := NewFreezeInfo(s.from, to, trxAmount)

	s.EveryBeforeFreeze(info)

	txID, err := free.Freeze(s.tronClient, s.from, s.privateKey, to, core.ResourceCode_ENERGY, trxAmount.IntPart()*TRXToSUN)

	s.EveryAfterFreeze(info, err)

	if err != nil {
		if _, saveErr := s.CreateFailRecord(to, err.Error(), trxAmount, true); saveErr != nil {
			s.logger.Error(`创建操作失败记录错误`, helpers.ZapError(err))
		}

		return errors.Wrap(err, `冻结`)
	}

	if _, err = s.CreateRecord(to, txID, trxAmount); err != nil {
		return err
	}

	return nil
}

/*FreezeForTransfer 用于TRC20的Transfer质押，质押足够的TRX
参数:
*	to   	string	参数1
返回值:
*	error	error 	返回值1
*/
func (s service) FreezeForTransfer(to string) error {
	return s.Freeze(to, decimal.New(TRXForSingleEnergy, 0))
}

/*UnFreeze 解冻
参数:
*	to   	string	收益地址
返回值:
*	error	error 	错误
*/
func (s service) UnFreeze(to string) error {
	now := helpers.Now()

	var (
		mutex redislock.Mutex
		err   error
		txID  string
	)
	// 加锁，需要是解锁时，是一次性解锁所有已经到期的质押
	if mutex, err = redislock.GetAndLock(s.locker, to, lockTimeout); err != nil {
		return errors.Wrap(err, `加锁`)
	}

	defer func() {
		_ = mutex.UnLock()
	}()

	info := NewFreezeInfo(s.from, to, decimal.Zero)

	s.EveryBeforeUnfreeze(info)

	txID, err = free.UnFreeze(s.tronClient, s.from, s.privateKey, to, core.ResourceCode_ENERGY)

	s.EveryAfterUnfreeze(info, err)

	if err != nil {
		if _, saveErr := s.CreateFailRecord(to, err.Error(), decimal.Zero, false); saveErr != nil {
			s.logger.Error(`创建操作失败记录错误`, helpers.ZapError(err))
		}

		return errors.Wrap(err, `解冻`)
	}

	return s.UpdateUnfreezeInfo(txID, now, helpers.Time(time.Unix(now.Unix(), 0).Add(time.Hour*-72).Unix()))
}

/*GetRecords 获取冻结记录
参数:
*	filter      	GetRecordFilter	过滤器
*	needAllCount	bool           	是否需要全部计数
返回值:
*	totalCount  	int64          	总数量，如果needAllCount==false, 这里返回0
*	records     	[]FreezeRecord 	符合条件的记录
*	err         	error          	错误
*/
func (s service) GetRecords(filter GetRecordFilter, needAllCount bool) (totalCount int64, records []FreezeRecord, err error) {
	return s.findRecords(filter, needAllCount)
}

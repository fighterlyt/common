package balance

import (
	"fmt"
	"github.com/fighterlyt/gotron-sdk/pkg/client"
	"github.com/fighterlyt/log"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.com/nova_dubai/common/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

// ErrAccountNotFound tron账户不存在的异常
var ErrAccountNotFound = errors.New("account not found")

// Service 余额服务
type Service struct {
	db                    *gorm.DB                                                   // 客户端
	engine                gin.IRouter                                                // http句柄
	checkBalanceInterval  time.Duration                                              // 查询余额间隔时间
	logger                log.Logger                                                 // 日志器
	tronClient            *client.GrpcClient                                         // tron客户端
	getBalanceFunc        func() (collectAddress, withdrawAddress string, err error) // 查询归集钱包和提现钱包的方法
	collectWalletBalance  *walletBalance                                             // 归集钱包余额
	withdrawWalletBalance *walletBalance                                             // 提款钱包余额
}

/*NewService 创建服务
参数:
*	db                  	*gorm.DB     	数据库连接
*	engine              	gin.IRouter  	http句柄
*	checkBalanceInterval	time.Duration	查询余额间隔时间
*	logger              	log.Logger   	日志器
返回值:
*	*Service            	*Service     	服务
*/
func NewService(db *gorm.DB, engine gin.IRouter, checkBalanceInterval time.Duration, logger log.Logger, getBalanceFunc func() (collectAddress, withdrawAddress string, err error)) *Service {
	service := &Service{db: db, engine: engine, checkBalanceInterval: checkBalanceInterval, logger: logger}

	return service
}

func (s *Service) start() (err error) {
	var collectAddress, withdrawAddress string

	for ; ; time.Sleep(s.checkBalanceInterval) {
		collectAddress, withdrawAddress, err = s.getBalanceFunc()
		if err != nil {
			s.logger.Error("查询归集钱包和提现钱包地址错误", zap.String("错误", err.Error()))

			continue
		}

		collectBalances, err := s.checkTrxAndUsdt(collectAddress)
		if err != nil {
			s.logger.Error("查询归集钱包余额失败", zap.String("错误", err.Error()))

			continue
		}

		s.collectWalletBalance.reset(collectBalances)

		withdrawBalances, err := s.checkTrxAndUsdt(withdrawAddress)
		if err != nil {
			s.logger.Error("查询提款钱包余额失败", zap.String("错误", err.Error()))

			continue
		}

		s.withdrawWalletBalance.reset(withdrawBalances)
	}
}

// 查询trx和usdt余额
func (s *Service) checkTrxAndUsdt(address string) (map[string]decimal.Decimal, error) {
	var (
		err                     error
		trxBalance, usdtBalance decimal.Decimal
	)

	trxBalance, err = s.checkBalance(address, model.TRX)
	if err != nil {
		return nil, err
	}

	usdtBalance, err = s.checkBalance(address, model.USDT)
	if err != nil {
		return nil, err
	}

	return map[string]decimal.Decimal{
		model.TRX:  trxBalance,
		model.USDT: usdtBalance,
	}, err
}

// CheckBalance 查询余额
func (s *Service) checkBalance(address, currency string) (decimal.Decimal, error) {
	switch currency {
	case model.TRX:
		account, err := s.tronClient.GetAccount(address)
		if err != nil {
			// 这个异常是tron账户未激活
			if errors.As(err, &ErrAccountNotFound) {
				return decimal.Zero, nil
			}

			return decimal.Zero, errors.Wrap(err, "获取tron余额失败")
		}

		return decimal.NewFromInt(account.Balance).Mul(decimal.New(1, -6)), nil
	default:
		contract, err := model.Trc20.ContractLocator().GetContract(currency)
		if err != nil {
			return decimal.Zero, errors.Wrapf(err, "获取合约[%s]地址失败", currency)
		}

		if contract == nil {
			return decimal.Zero, fmt.Errorf("不支持的代币[%s]", currency)
		}

		balance, err := s.tronClient.TRC20ContractBalance(address, contract.Address())
		if err != nil {
			return decimal.Zero, errors.Wrap(err, "获取代币余额失败")
		}

		// 小数点位数
		decimals, err := s.tronClient.TRC20GetDecimals(contract.Address())
		if err != nil {
			return decimal.Zero, errors.Wrap(err, "获取代币余额失败")
		}

		return decimal.NewFromBigInt(balance, -int32(decimals.Int64())), nil
	}
}

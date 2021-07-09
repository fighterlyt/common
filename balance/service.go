package balance

import (
	"fmt"
	"github.com/fighterlyt/gotron-sdk/pkg/client"
	"github.com/fighterlyt/log"
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
	checkBalanceInterval  time.Duration                                              // 查询余额间隔时间
	logger                log.Logger                                                 // 日志器
	tronClient            *client.GrpcClient                                         // tron客户端
	getBalanceFunc        func() (collectAddress, withdrawAddress string, err error) // 查询归集钱包和提现钱包的方法
	collectWalletBalance  *walletBalance                                             // 归集钱包余额
	withdrawWalletBalance *walletBalance                                             // 提款钱包余额
	collectAddress        string                                                     // 归集钱包地址
	withdrawAddress       string                                                     // 提现钱包地址
}

func NewService(db *gorm.DB, checkBalanceInterval time.Duration, logger log.Logger, getBalanceFunc func() (collectAddress, withdrawAddress string, err error)) *Service {
	service := &Service{db: db, checkBalanceInterval: checkBalanceInterval, logger: logger, getBalanceFunc: getBalanceFunc}

	service.start()

	return service
}

func (s *Service) start() {
	var err error

	for ; ; time.Sleep(s.checkBalanceInterval) {
		s.collectAddress, s.withdrawAddress, err = s.getBalanceFunc()
		if err != nil {
			s.logger.Error("查询归集钱包和提现钱包地址错误", zap.String("错误", err.Error()))

			continue
		}

		collectBalances, err := s.checkTrxAndUsdt(s.collectAddress)
		if err != nil {
			s.logger.Error("查询归集钱包余额失败", zap.String("错误", err.Error()))

			continue
		}

		s.collectWalletBalance.reset(collectBalances)

		withdrawBalances, err := s.checkTrxAndUsdt(s.withdrawAddress)
		if err != nil {
			s.logger.Error("查询提款钱包余额失败", zap.String("错误", err.Error()))

			continue
		}

		s.withdrawWalletBalance.reset(withdrawBalances)
	}
}

func (s *Service) GetWalletBalance() *WalletBalanceInfo {
	return &WalletBalanceInfo{
		CollectWalletInfo: WalletInfo{
			Address:  s.collectAddress,
			Balances: s.collectWalletBalance.get(),
		},
		WithdrawWalletInfo: WalletInfo{
			Address:  s.withdrawAddress,
			Balances: s.withdrawWalletBalance.get(),
		},
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

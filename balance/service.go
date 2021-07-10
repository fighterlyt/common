package balance

import (
	"fmt"
	"time"

	"github.com/fighterlyt/gotron-sdk/pkg/client"
	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.com/nova_dubai/common/helpers"
	"gitlab.com/nova_dubai/common/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
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
	walletMetrics         *metrics                                                   // 监控信息
	currency              string                                                     // 查询的币种
}

func NewService(db *gorm.DB, tronClient *client.GrpcClient, currency string, checkBalanceInterval time.Duration, logger log.Logger, getBalanceFunc func() (collectAddress, withdrawAddress string, err error)) (*Service, error) { // nolint:golint,lll
	walletMetrics, err := newMetrics(logger)
	if err != nil {
		return nil, errors.Wrap(err, "启动监控失败")
	}

	service := &Service{
		db:                    db,
		tronClient:            tronClient,
		checkBalanceInterval:  checkBalanceInterval,
		logger:                logger,
		getBalanceFunc:        getBalanceFunc,
		walletMetrics:         walletMetrics,
		currency:              currency,
		collectWalletBalance:  newWalletBalance(),
		withdrawWalletBalance: newWalletBalance(),
	}

	helpers.EnsureGo(logger, func() {
		service.start()
	})

	return service, nil
}

func (s *Service) start() {
	var (
		err                               error
		collectBalances, withdrawBalances map[string]decimal.Decimal
	)

	for ; ; time.Sleep(s.checkBalanceInterval) {
		s.collectAddress, s.withdrawAddress, err = s.getBalanceFunc()
		if err != nil {
			s.logger.Error("查询归集钱包和提现钱包地址错误", zap.String("错误", err.Error()))

			continue
		}

		collectBalances, err = s.checkTrxAndUsdt(s.collectAddress)
		if err != nil {
			s.logger.Error("查询归集钱包余额失败", zap.String("错误", err.Error()))

			continue
		}

		// 保存监控数据
		s.walletMetrics.collectWalletBalance.WithLabelValuesSet(s.getBalanceByCurrency(collectBalances, model.TRX), model.TRX)
		s.walletMetrics.collectWalletBalance.WithLabelValuesSet(s.getBalanceByCurrency(collectBalances, model.USDT), model.USDT)

		s.collectWalletBalance.reset(collectBalances)

		withdrawBalances, err = s.checkTrxAndUsdt(s.withdrawAddress)
		if err != nil {
			s.logger.Error("查询提款钱包余额失败", zap.String("错误", err.Error()))

			continue
		}

		// 保存监控数据
		s.walletMetrics.withdrawWalletBalance.WithLabelValuesSet(s.getBalanceByCurrency(withdrawBalances, model.TRX), model.TRX)
		s.walletMetrics.withdrawWalletBalance.WithLabelValuesSet(s.getBalanceByCurrency(withdrawBalances, model.USDT), model.USDT)

		s.withdrawWalletBalance.reset(withdrawBalances)
	}
}

func (s Service) getBalanceByCurrency(balances map[string]decimal.Decimal, currency string) float64 {
	balance, _ := balances[currency].Float64()

	return balance
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

	usdtBalance, err = s.checkBalance(address, s.currency)
	if err != nil {
		return nil, err
	}

	return map[string]decimal.Decimal{
		model.TRX:  trxBalance,
		s.currency: usdtBalance,
	}, err
}

// CheckBalance 查询余额
func (s *Service) checkBalance(address, currency string) (decimal.Decimal, error) {
	switch currency {
	case model.TRX:
		return s.checkTrxBalance(address)
	default:
		return s.checkContractBalance(address, currency)
	}
}

// 查询trx余额
func (s Service) checkTrxBalance(address string) (decimal.Decimal, error) {
	account, err := s.tronClient.GetAccount(address)
	if err != nil {
		// 这个异常是tron账户未激活
		if errors.As(err, &ErrAccountNotFound) {
			return decimal.Zero, nil
		}

		return decimal.Zero, errors.Wrap(err, "获取tron余额失败")
	}

	return decimal.NewFromInt(account.Balance).Mul(decimal.New(1, -6)), nil
}

// 查询合约币种余额
func (s Service) checkContractBalance(address, currency string) (decimal.Decimal, error) {
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

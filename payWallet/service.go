package paywallet

import (
	"gitlab.com/nova_dubai/common/parameters"
	"go.uber.org/zap"
	"time"

	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.com/nova_dubai/common/helpers"
	"gitlab.com/nova_dubai/common/model"
)

// Service 出款钱包余额服务
type Service struct {
	checkBalanceInterval time.Duration                                                                    // 查询余额间隔时间
	logger               log.Logger                                                                       // 日志器
	CheckBalanceFunc     func(protocol model.Protocol, address, currency string) (decimal.Decimal, error) // 查询余额的方法
	parameterService     parameters.Service
	tronArgument         model.WalletArgument
	ethArgument          model.WalletArgument
	tronBalanceDetail    *model.BalanceDetail
	ethBalanceDetail     *model.BalanceDetail
}

func NewService(checkBalanceInterval time.Duration, parameterService parameters.Service, tronArgument, ethArgument model.WalletArgument, CheckBalanceFunc func(protocol model.Protocol, address, currency string) (decimal.Decimal, error), logger log.Logger) (*Service, error) { // nolint:golint,lll
	service := &Service{
		checkBalanceInterval: checkBalanceInterval,
		logger:               logger,
		CheckBalanceFunc:     CheckBalanceFunc,
		tronArgument:         tronArgument,
		ethArgument:          ethArgument,
		parameterService:     parameterService,
		tronBalanceDetail:    &model.BalanceDetail{},
		ethBalanceDetail:     &model.BalanceDetail{},
	}

	helpers.EnsureGo(logger, func() {
		service.start()
	})

	return service, nil
}

func (s *Service) start() {
	for ; ; time.Sleep(s.checkBalanceInterval) {
		s.singleCheck()
	}
}

// 单次查询余额
func (s *Service) singleCheck() {
	if err := s.checkWalletBalance(s.tronArgument); err != nil {
		s.logger.Error("查询钱包余额失败", zap.Any("参数", s.tronArgument), zap.String("错误", err.Error()))
	}

	if err := s.checkWalletBalance(s.ethArgument); err != nil {
		s.logger.Error("查询钱包余额失败", zap.Any("参数", s.tronArgument), zap.String("错误", err.Error()))
	}
}

func (s *Service) checkWalletBalance(argument model.WalletArgument) (err error) {
	result := model.BalanceDetail{Protocol: argument.Protocol}

	result.Address, err = s.parameterService.GetString(argument.WalletKey)
	if err != nil {
		return errors.Wrap(err, "查询业务参数失败")
	}

	// 查询usdt余额
	result.USDTBalance, err = s.CheckBalanceFunc(argument.Protocol, result.Address, argument.Symbol)
	if err != nil {
		return errors.Wrap(err, "查询USDT余额失败")
	}

	// 查询手续费余额
	feeSymbol := model.ETH
	if argument.Protocol == model.Trc20 {
		feeSymbol = model.TRX
	}

	result.FeeBalance, err = s.CheckBalanceFunc(argument.Protocol, result.Address, feeSymbol)
	if err != nil {
		return errors.Wrap(err, "查询USDT余额失败")
	}

	return nil
}

// GetPayWalletDetails 查询钱包余额
func (s Service) GetPayWalletDetails() (result []model.BalanceDetail, err error) {
	return result, nil
}

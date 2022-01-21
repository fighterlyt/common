package paywallet

import (
	"fmt"
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
	getAddressFunc       func(protocol model.Protocol) (address string, symbol string, err error)         // 获取参数的方法
	protocols            []model.Protocol
	tronBalanceDetail    *model.BalanceDetail
	ethBalanceDetail     *model.BalanceDetail
}

func NewService(protocols []model.Protocol, checkBalanceInterval time.Duration, getAddressFunc func(protocol model.Protocol) (address string, symbol string, err error), CheckBalanceFunc func(protocol model.Protocol, address, currency string) (decimal.Decimal, error), logger log.Logger) (*Service, error) { // nolint:golint,lll
	service := &Service{
		protocols:            protocols,
		checkBalanceInterval: checkBalanceInterval,
		logger:               logger,
		CheckBalanceFunc:     CheckBalanceFunc,
		getAddressFunc:       getAddressFunc,
		tronBalanceDetail:    &model.BalanceDetail{Protocol: model.Trc20},
		ethBalanceDetail:     &model.BalanceDetail{Protocol: model.Erc20},
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
	for _, protocol := range s.protocols {
		address, symbol, err := s.getAddressFunc(protocol)
		if err != nil {
			s.logger.Error("查询钱包余额失败", zap.Any("协议", protocol), zap.String("错误", err.Error()))

			continue
		}

		if err = s.checkWalletBalance(protocol, address, symbol); err != nil {
			s.logger.Error("查询钱包余额失败", zap.Strings("协议/地址/币种", []string{protocol.String(), address, symbol}), zap.String("错误", err.Error()))
		}
	}
}

func (s *Service) checkWalletBalance(protocol model.Protocol, address, symbol string) (err error) {
	var result *model.BalanceDetail

	switch protocol {
	case model.Trc20:
		result = s.tronBalanceDetail
	case model.Erc20:
		result = s.ethBalanceDetail
	default:
		return fmt.Errorf("不支持的协议[%s]", protocol)
	}

	result.Address = address

	// 查询usdt余额
	result.USDTBalance, err = s.CheckBalanceFunc(protocol, address, symbol)
	if err != nil {
		return errors.Wrap(err, "查询USDT余额失败")
	}

	// 查询手续费余额
	feeSymbol := model.ETH
	if protocol == model.Trc20 {
		feeSymbol = model.TRX
	}

	result.FeeBalance, err = s.CheckBalanceFunc(protocol, address, feeSymbol)
	if err != nil {
		return errors.Wrap(err, "查询手续费余额失败")
	}

	return nil
}

// GetPayWalletDetails 查询钱包余额
func (s Service) GetPayWalletDetails() (result []*model.BalanceDetail, err error) {
	if s.tronBalanceDetail.Address != "" {
		result = append(result, s.tronBalanceDetail)
	}

	if s.ethBalanceDetail.Address != "" {
		result = append(result, s.ethBalanceDetail)
	}

	return result, nil
}

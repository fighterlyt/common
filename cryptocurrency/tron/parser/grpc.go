package parser

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fighterlyt/common/cryptocurrency"
	"github.com/fighterlyt/common/model"
	"github.com/fighterlyt/gotron-sdk/pkg/client"
	"github.com/fighterlyt/gotron-sdk/pkg/common"
	"github.com/fighterlyt/gotron-sdk/pkg/proto/api"
	"github.com/fighterlyt/gotron-sdk/pkg/proto/core"
	"github.com/fighterlyt/log"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"go.uber.org/multierr"

	"github.com/fighterlyt/common/helpers"
	"go.uber.org/zap"
)

var (
	abi trc20Abi
)

const (
	kilo    = 1000
	success = "SUCCESS"
)

type grpcParser struct {
	grpcClient *client.GrpcClient         // grpc客户端
	logger     log.Logger                 // 日志器
	concern    cryptocurrency.Concern     // concern 对地址的关注
	contract   cryptocurrency.Contract    // 关注的智能合约
	notify     cryptocurrency.TradeNotify // 交易通知
	includeTRX bool                       // 是否包含TRX
}

/*
NewGRPCTronScanParser grpc解析器
参数:
*	concern 					cryptocurrency.Concern			关心账号
*	grpcClient  				*client.GrpcClient		        波场grpc客户端
*	logger  					log.Logger					    日志器
*	contract					cryptocurrency.Contract		    合约
*	notify  					model.TradeDetailNotify		    通知器
返回值:
*	model.TronParser	model.TronParser					解析器
*/
func NewGRPCTronScanParser(concern cryptocurrency.Concern, grpClient *client.GrpcClient, logger log.Logger, contract cryptocurrency.Contract, notify cryptocurrency.TradeNotify) TronParser { //nolint:golint,lll
	return &grpcParser{
		grpcClient: grpClient,
		concern:    concern,
		logger:     logger,
		contract:   contract,
		notify:     notify,
	}
}

func (g *grpcParser) IncludeTRX(include bool) {
	g.includeTRX = include
}

/*
IsBlockConfirmed 区块是否被确认
参数:
*	ctx        	context.Context		上下文
*	blockNumber	int64				区块号
返回值:
*	confirmed	bool			    是否被确认
*	err      	error			    错误信息
*/
func (g grpcParser) IsBlockConfirmed(_ context.Context, _ int64) (confirmed bool, err error) {
	return true, nil
}

/*
Parse 区块解析
参数:
*	ctx        	context.Context		上下文
*	blockNumber	int64				区块号
返回值:
*	trades		[]*model.Trade		交易详情
*	err   		error				错误信息
*/
func (g grpcParser) Parse(ctx context.Context, blockNumber int64) (trades []*cryptocurrency.Trade, err error) {
	logger := g.logger.With(zap.Int64(`区块号码`, blockNumber))

	var (
		blockExtension *api.BlockExtention
		notifyDetails  []*cryptocurrency.TransactionDetail
		parseErr       error
		trade          *cryptocurrency.Trade
		detail         *cryptocurrency.TransactionDetail
	)

	logger.Info(`解析区块`)

	if blockExtension, err = g.grpcClient.GetBlockByNum(blockNumber); err != nil {
		helpers.IgnoreError(g.logger, "重启波场grpc客户端", func() error {
			return g.grpcClient.Reconnect(g.grpcClient.Address)
		})

		return nil, errors.Wrap(err, "GetBlockByName")
	}

	for _, tx := range blockExtension.GetTransactions() {
		if trade, detail, parseErr = g.ParseTX(ctx, tx, logger, blockNumber); parseErr == nil && trade != nil {
			trades = append(trades, trade)
			notifyDetails = append(notifyDetails, detail)
		} else if parseErr != nil {
			err = multierr.Append(err, parseErr)
			logger.Error(`解析错误`, zap.String(`错误`, parseErr.Error()))
		}
	}

	logger.Info(`准备通知交易`, zap.Int(`交易数量`, len(notifyDetails)), zap.Bool(`是否有通知项`, g.notify != nil))

	if len(notifyDetails) > 0 && g.notify != nil {
		helpers.IgnoreError(g.logger, "通知交易", func() error {
			return g.notify.Notify(cryptocurrency.Trc20, notifyDetails)
		})
	}

	return trades, err
}

func getContract(tx *api.TransactionExtention) *core.Transaction_Contract {
	if contract := tx.GetTransaction().GetRawData().GetContract(); len(contract) != 0 {
		return contract[0]
	}

	return nil
}

/*
ParseTX  解析交易
参数:
*	_          	context.Context                  	上下文
*	tx         	*api.TransactionExtention        	交易
*	logger     	log.Logger                       	日志器
*	blockNumber	int64                            	区块号
返回值:
*	trade      	*cryptocurrency.Trade            	返回值1
*	detail     	*cryptocurrency.TransactionDetail	返回值2
*	err        	error                            	返回值3
*/
func (g grpcParser) ParseTX(_ context.Context, tx *api.TransactionExtention, logger log.Logger, blockNumber int64) (trade *cryptocurrency.Trade, detail *cryptocurrency.TransactionDetail, err error) { //nolint:lll
	var (
		contract *core.Transaction_Contract
	)

	txID := hexutil.Encode(tx.Txid)[2:]

	// 非智能合约或者不是触发了只能合约
	if contract = getContract(tx); contract == nil {
		logger.Info(`非智能合约或者不是触发智能合约`)
		return nil, nil, nil
	}

	switch contract.GetType() {
	case core.Transaction_Contract_TriggerSmartContract:
		return g.parseTrc20(contract, tx, logger, blockNumber, txID)
	case core.Transaction_Contract_TransferContract:
		if !g.includeTRX {
			return nil, nil, nil
		}

		return g.parseTRx(contract, tx, logger, blockNumber, txID)
	default:
		return nil, nil, nil
	}
}

func (g grpcParser) parseTRx(contract *core.Transaction_Contract, tx *api.TransactionExtention, logger log.Logger, blockNumber int64, txID string) (trade *cryptocurrency.Trade, detail *cryptocurrency.TransactionDetail, err error) {
	var (
		matched   bool
		toAddress string
		info      *core.TransactionInfo
	)

	transaction := &core.TransferContract{}

	if err = proto.Unmarshal(contract.Parameter.GetValue(), transaction); err != nil {
		return nil, nil, errors.Wrap(err, `解析value`)
	}

	ret := tx.Transaction.GetRet()

	if ret == nil || !tx.GetResult().Result || core.Transaction_ResultContractResult_name[int32(ret[0].ContractRet)] != success { // nolint:golint,lll
		return nil, nil, nil
	}

	toAddress = common.EncodeCheck(transaction.ToAddress)

	amount := decimal.New(transaction.Amount, -g.contract.Precision())

	ownerAddress := common.EncodeCheck(transaction.OwnerAddress)

	logger.Info(`判断交易是否符合条件`, zap.Strings(`ownAddress/toAddress/amount`, []string{ownerAddress, toAddress, amount.String()}))

	if g.concern == nil {
		return nil, nil, nil
	}

	if matched, _, err = g.concern.FilterConcernedAccounts(ownerAddress, toAddress, amount); err != nil {
		return nil, nil, errors.Wrapf(err, `判断关注交易错误,转出[%s]转入[%s],金额[%s]`, ownerAddress, toAddress, amount.String())
	}

	if !matched {
		logger.Info(`不匹配`, zap.Strings(`ownAddress/toAddress/amount`, []string{ownerAddress, toAddress, amount.String()}))
		return nil, nil, nil
	}

	logger.Info("交易匹配")

	if info, err = g.grpcClient.GetTransactionInfoByID(hexutil.Encode(tx.Txid)); err != nil {
		helpers.IgnoreError(g.logger, "重启波场grpc客户端", func() error {
			return g.grpcClient.Reconnect(g.grpcClient.Address)
		})

		return nil, nil, errors.Wrap(err, "GetTransactionInfoByID")
	}

	fee := decimal.New(info.GetFee(), -6)

	tradeTime := info.BlockTimeStamp / int64(kilo)

	trade = cryptocurrency.NewTrade(cryptocurrency.Trc20, ownerAddress, toAddress, amount, model.TRX, txID, tradeTime, blockNumber, fee, cryptocurrency.TradeTransfer) // nolint:golint,lll

	detail = cryptocurrency.NewFullTransactionDetail(amount, cryptocurrency.Trc20, model.TRX, ownerAddress, toAddress, blockNumber, txID, fee, tradeTime, cryptocurrency.TradeTransfer) //nolint:lll

	return trade, detail, nil
}

func (g grpcParser) parseTrc20(contract *core.Transaction_Contract, tx *api.TransactionExtention, logger log.Logger, blockNumber int64, txID string) (trade *cryptocurrency.Trade, detail *cryptocurrency.TransactionDetail, err error) {
	var (
		matched   bool
		toAddress string
		value     int64
		info      *core.TransactionInfo
		tradeKind cryptocurrency.TradeKind
	)

	transaction := &core.TriggerSmartContract{}

	if err = proto.Unmarshal(contract.Parameter.GetValue(), transaction); err != nil {
		return nil, nil, errors.Wrap(err, `解析value`)
	}

	if common.EncodeCheck(transaction.ContractAddress) != g.contract.Address() {
		return nil, nil, nil
	}

	methodType := abi.MethodType(strings.TrimPrefix(hexutil.Encode(transaction.Data), "0x"))
	ret := tx.Transaction.GetRet()

	if ret == nil || !tx.GetResult().Result || core.Transaction_ResultContractResult_name[int32(ret[0].ContractRet)] != success { // nolint:golint,lll
		return nil, nil, nil
	}

	var (
		method func(string) (string, int64, error)
	)

	switch methodType {
	case trc20Transfer:
		method = abi.UnpackTransfer

		tradeKind = cryptocurrency.TradeTransfer
	case trc20Approve:
		method = abi.UnpackApprove

		tradeKind = cryptocurrency.TradeApprove
	case trc20TransferFrom:
		method = abi.UnpackTransferFrom

		tradeKind = cryptocurrency.TradeTransfer
	default:
		return nil, nil, nil
	}

	if toAddress, value, err = method(strings.TrimPrefix(hexutil.Encode(transaction.Data), "0x")); err != nil {
		return nil, nil, errors.Wrapf(err, "解析trc20交易[%s]数据[%s]", hexutil.Encode(tx.Txid)[2:], hexutil.Encode(transaction.Data))
	}

	amount := decimal.New(value, -g.contract.Precision())

	ownerAddress := common.EncodeCheck(transaction.OwnerAddress)

	logger.Info(`判断交易是否符合条件`, zap.Strings(`ownAddress/toAddress/amount`, []string{ownerAddress, toAddress, amount.String()}))

	if g.concern == nil {
		return nil, nil, nil
	}

	if matched, _, err = g.concern.FilterConcernedAccounts(ownerAddress, toAddress, amount); err != nil {
		return nil, nil, errors.Wrapf(err, `判断关注交易错误,转出[%s]转入[%s],金额[%s]`, ownerAddress, toAddress, amount.String())
	}

	if !matched {
		logger.Info(`不匹配`, zap.Strings(`ownAddress/toAddress/amount`, []string{ownerAddress, toAddress, amount.String()}))
		return nil, nil, nil
	}

	logger.Info("交易匹配")

	if info, err = g.grpcClient.GetTransactionInfoByID(hexutil.Encode(tx.Txid)); err != nil {
		helpers.IgnoreError(g.logger, "重启波场grpc客户端", func() error {
			return g.grpcClient.Reconnect(g.grpcClient.Address)
		})

		return nil, nil, errors.Wrap(err, "GetTransactionInfoByID")
	}

	fee := decimal.New(info.GetFee(), -6)

	tradeTime := info.BlockTimeStamp / int64(kilo)

	trade = cryptocurrency.NewTrade(cryptocurrency.Trc20, ownerAddress, toAddress, amount, g.contract.Token(), txID, tradeTime, blockNumber, fee, tradeKind) // nolint:golint,lll

	detail = cryptocurrency.NewFullTransactionDetail(amount, cryptocurrency.Trc20, g.contract.Token(), ownerAddress, toAddress, blockNumber, txID, fee, tradeTime, tradeKind) //nolint:lll

	return trade, detail, nil
}

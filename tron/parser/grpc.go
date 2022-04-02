package parser

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fighterlyt/gotron-sdk/pkg/client"
	"github.com/fighterlyt/gotron-sdk/pkg/common"
	"github.com/fighterlyt/gotron-sdk/pkg/proto/api"
	"github.com/fighterlyt/gotron-sdk/pkg/proto/core"
	"github.com/fighterlyt/log"
	"github.com/golang/protobuf/proto" // nolint:golint,staticcheck
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.com/nova_dubai/usdtpay/model"
	"go.uber.org/multierr"

	"gitlab.com/nova_dubai/usdtpay/helpers"
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
	grpcClient *client.GrpcClient // grpc客户端
	logger     log.Logger         // 日志器
	concern    model.Concern      // concern 对地址的关注
	contract   model.Contract     // 关注的智能合约
	notify     model.TradeNotify  // 交易通知
}

/*NewGRPCTronScanParser grpc解析器
参数:
*	concern 					model.Concern				关心账号
*	grpcClient  				*client.GrpcClient		    波场grpc客户端
*	logger  					log.Logger					日志器
*	contract					model.Contract				合约
*	notify  					model.TradeDetailNotify		通知器
返回值:
*	model.TronParser	model.TronParser					解析器
*/
func NewGRPCTronScanParser(concern model.Concern, grpClient *client.GrpcClient, logger log.Logger, contract model.Contract, notify model.TradeNotify) model.TronParser { //nolint:golint,lll
	return &grpcParser{
		grpcClient: grpClient,
		concern:    concern,
		logger:     logger,
		contract:   contract,
		notify:     notify,
	}
}

/*IsBlockConfirmed 区块是否被确认
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

/*Parse 区块解析
参数:
*	ctx        	context.Context		上下文
*	blockNumber	int64				区块号
返回值:
*	trades		[]*model.Trade		交易详情
*	err   		error				错误信息
*/
func (g grpcParser) Parse(_ context.Context, blockNumber int64) (trades []*model.Trade, err error) {
	logger := g.logger.With(zap.Int64(`区块号码`, blockNumber))

	var (
		blockExtension *api.BlockExtention
	)

	logger.Info(`解析区块`)

	if blockExtension, err = g.grpcClient.GetBlockByNum(blockNumber); err != nil {
		helpers.IgnoreError(g.logger, "重启波场grpc客户端", func() error {
			return g.grpcClient.Reconnect(g.grpcClient.Address)
		})

		return nil, errors.Wrap(err, "GetBlockByName")
	}

	var (
		txes          = blockExtension.GetTransactions()
		matched       bool
		toAddress     string
		value         int64
		notifyDetails []*model.TransactionDetail
		contract      *core.Transaction_Contract
		parseErr      error
		info          *core.TransactionInfo
		tradeKind     model.TradeKind
	)

	for _, tx := range txes {
		if parseErr != nil {
			err = multierr.Append(err, parseErr)
			logger.Error(`解析错误`, zap.String(`错误`, parseErr.Error()))

			parseErr = nil
		}
		// 非智能合约或者不是触发了只能合约
		if contract = getContract(tx); contract == nil || contract.GetType() != core.Transaction_Contract_TriggerSmartContract {
			logger.Info(`非智能合约或者不是触发智能合约`)
			continue
		}

		transaction := &core.TriggerSmartContract{}
		if parseErr = proto.Unmarshal(contract.Parameter.GetValue(), transaction); parseErr != nil {
			err = errors.Wrap(parseErr, `解析value`)
			continue
		}

		if common.EncodeCheck(transaction.ContractAddress) != g.contract.Address() {
			continue
		}

		methodType := abi.MethodType(strings.TrimPrefix(hexutil.Encode(transaction.Data), "0x"))
		ret := tx.Transaction.GetRet()

		if ret == nil {
			logger.Info("不符合要求", zap.Any("原因", "交易ret为空"))

			continue
		}

		if !tx.GetResult().Result || core.Transaction_ResultContractResult_name[int32(ret[0].ContractRet)] != success { // nolint:golint,lll
			logger.Info("不符合要求", zap.Any("methodType", methodType), zap.Any("交易", string(transaction.Data)))

			continue
		}

		switch methodType {
		case trc20Transfer:
			if toAddress, value, parseErr = abi.UnpackTransfer(strings.TrimPrefix(hexutil.Encode(transaction.Data), "0x")); parseErr != nil {
				parseErr = errors.Wrapf(parseErr, "解析trc20交易[%s]数据[%s]", hexutil.Encode(tx.Txid)[2:], hexutil.Encode(transaction.Data))

				continue
			}

			tradeKind = model.TradeTransfer
		case trc20Approve:
			if toAddress, value, parseErr = abi.UnpackApprove(strings.TrimPrefix(hexutil.Encode(transaction.Data), "0x")); parseErr != nil {
				parseErr = errors.Wrapf(parseErr, "解析trc20交易[%s]数据[%s]", hexutil.Encode(tx.Txid)[2:], hexutil.Encode(transaction.Data))

				continue
			}

			tradeKind = model.TradeApprove
		case trc20TransferFrom:
			if toAddress, value, parseErr = abi.UnpackTransferFrom(strings.TrimPrefix(hexutil.Encode(transaction.Data), "0x")); parseErr != nil {
				parseErr = errors.Wrapf(parseErr, "解析trc20交易[%s]数据[%s]", hexutil.Encode(tx.Txid)[2:], hexutil.Encode(transaction.Data))

				continue
			}

			tradeKind = model.TradeTransfer
		default:
			logger.Info(`不支持的方法`, zap.Any("methodType", methodType), zap.Any("交易", string(transaction.Data)))
			continue
		}

		amount := decimal.New(value, -g.contract.Precision())

		ownerAddress := common.EncodeCheck(transaction.OwnerAddress)

		logger.Info(`判断交易是否符合条件`, zap.Strings(`ownAddress/toAddress/amount`, []string{ownerAddress, toAddress, amount.String()}))

		if g.concern == nil {
			continue
		}

		if matched, _, parseErr = g.concern.FilterConcernedAccounts(ownerAddress, toAddress, amount); parseErr != nil {
			parseErr = errors.Wrapf(parseErr, `判断关注交易错误,转出[%s]转入[%s],金额[%s]`, ownerAddress, toAddress, amount.String())

			continue
		}

		if !matched {
			logger.Info(`不匹配`, zap.Strings(`ownAddress/toAddress/amount`, []string{ownerAddress, toAddress, amount.String()}))
			continue
		}

		logger.Info("交易匹配")

		if info, parseErr = g.grpcClient.GetTransactionInfoByID(hexutil.Encode(tx.Txid)); parseErr != nil {
			helpers.IgnoreError(g.logger, "重启波场grpc客户端", func() error {
				return g.grpcClient.Reconnect(g.grpcClient.Address)
			})

			parseErr = errors.Wrap(parseErr, "GetTransactionInfoByID")

			continue
		}

		fee := decimal.New(info.GetFee(), -6)

		txID := hexutil.Encode(tx.Txid)[2:]
		tradeTime := info.BlockTimeStamp / int64(kilo)
		trades = append(trades, model.NewTrade(model.Trc20, ownerAddress, toAddress, amount, g.contract.Token(), txID, tradeTime, blockNumber, fee, tradeKind))

		detail := model.NewFullTransactionDetail(amount, model.Trc20, g.contract.Token(), ownerAddress, toAddress, blockNumber, txID, fee, tradeTime, tradeKind) //nolint:lll
		notifyDetails = append(notifyDetails, detail)
	}

	if parseErr != nil {
		err = multierr.Append(err, parseErr)
	}

	logger.Info(`准备通知交易`, zap.Int(`交易数量`, len(notifyDetails)), zap.Bool(`是否有通知项`, g.notify != nil))

	if len(notifyDetails) > 0 && g.notify != nil {

		helpers.IgnoreError(g.logger, "通知交易", func() error {
			return g.notify.Notify(model.Trc20, notifyDetails)
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

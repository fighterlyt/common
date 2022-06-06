package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"flag"
	"math/big"
	"os"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
	tronAddress "github.com/fighterlyt/gotron-sdk/pkg/address"
	"github.com/fighterlyt/gotron-sdk/pkg/client"
	"github.com/fighterlyt/gotron-sdk/pkg/client/transaction"
	"github.com/fighterlyt/gotron-sdk/pkg/common"
	"github.com/fighterlyt/gotron-sdk/pkg/keystore"
	"github.com/fighterlyt/gotron-sdk/pkg/proto/api"
	"github.com/fighterlyt/gotron-sdk/pkg/proto/core"
	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// 将指定钱包里的TRX/USDT 转到 指定钱包

var (
	filePath  = `user_wallet.json`
	USDT      = `TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t`
	USDC      = `TEkxiTehnzSmSe2XqrBj4w32RUN966rdz8`
	TRX       = ``
	logger    log.Logger
	operation = ``
)

func main() {
	flag.StringVar(&operation, `op`, ``, `操作`)
	flag.Parse()

	logger, _ = log.NewEasyLogger(true, false, ``, `test`)

	var (
		err error
	)

	switch operation {
	case `1`:
		logger.Info(`查询有余额的钱包`)

		err = first()
	case `2`:
		logger.Info(`过滤统计`)

		err = second()
	case `3`:
		logger.Info(`转账`)

		err = transfer(`count.json`, `TW1SWgnwRJPDKzkw2ZZJBsNF7vkLyMU4Mk`)
	default:
		logger.Info(`无操作`)
	}

	if err != nil {
		panic(operation + err.Error())
	}
}

func second() error {
	var (
		file      *os.File
		err       error
		balances  = map[string]Balance{}
		usdtTotal decimal.Decimal
		trxTotal  decimal.Decimal
		usdcTotal decimal.Decimal
	)

	if file, err = os.Open(`output.json`); err != nil {
		return errors.Wrap(err, `打开`)
	}

	defer func() {
		_ = file.Close()
	}()

	if err = json.NewDecoder(file).Decode(&balances); err != nil {
		return errors.Wrap(err, `解码`)
	}

	result := make(map[string]Balance, 100)

	for private, balance := range balances {
		if balance.TRX.LessThan(decimal.New(2, 0)) && balance.USDT.LessThan(decimal.Zero) && balance.USDC.LessThanOrEqual(decimal.New(2, 0)) {
			continue
		}

		usdtTotal = usdtTotal.Add(balance.USDT)
		trxTotal = trxTotal.Add(balance.TRX)
		usdcTotal = usdcTotal.Add(balance.USDC)

		result[private] = balances[private]
	}

	if err = output(result, `count.json`); err != nil {
		return errors.Wrap(err, `输出`)
	}

	logger.Info(`汇总完成`, zap.Int(`钱包数量`, len(result)), zap.Strings(`USDT/TRX/USDC`, []string{usdtTotal.String(), trxTotal.String(), usdcTotal.String()})) //nolint:lll

	return nil
}

/*first 从钱包json文件中逐个查询余额，过滤后记录
参数:
返回值:
*	error	error	返回值1
*/
func first() error {
	var (
		nonZero   map[string]Balance
		err       error
		addresses map[string]string
	)

	if addresses, err = load(); err != nil {
		return errors.Wrap(err, `加载数据`)
	}

	if nonZero, err = checkBalance(addresses); err != nil {
		return errors.Wrap(err, `过滤`)
	}

	for _, balance := range nonZero {
		logger.Info(`余额不为0`, zap.String(`地址`, balance.Address), zap.String(`USDT`, balance.USDT.String()), zap.String(`TRX`, balance.TRX.String())) //nolint:lll
	}

	if err = output(nonZero, `output.json`); err != nil {
		return errors.Wrap(err, `保存`)
	}

	return nil
}

func output(data interface{}, fileName string) error {
	var (
		file *os.File
		err  error
	)

	if file, err = os.Create(fileName); err != nil {
		return errors.Wrap(err, `创建文件`)
	}

	defer func() {
		_ = file.Close()
	}()

	if err = json.NewEncoder(file).Encode(data); err != nil {
		return errors.Wrap(err, `写入`)
	}

	return nil
}

type Balance struct {
	Address string
	USDT    decimal.Decimal
	TRX     decimal.Decimal
	USDC    decimal.Decimal
}

func getUSDTBalance(address, token string, client *client.GrpcClient) (balance decimal.Decimal, err error) {
	var (
		money *big.Int
	)

	for {
		if money, err = client.TRC20ContractBalance(address, token); err != nil && !strings.Contains(err.Error(), `exceeded`) {
			return decimal.Zero, errors.Wrapf(err, `判断USDT余额[%s]`, address)
		}

		if err == nil {
			break
		}
	}

	return decimal.NewFromBigInt(money, -6), nil
}

func getTRXBalance(address string, client *client.GrpcClient) (balance decimal.Decimal, err error) {
	var (
		account *core.Account
	)

	for {
		account, err = client.GetAccount(address)

		if err == nil {
			return decimal.New(account.Balance, -6), nil
		}

		if strings.Contains(err.Error(), `account not found`) {
			return decimal.Zero, nil
		}

		if strings.Contains(err.Error(), `exceeded`) {
			continue
		}

		return decimal.Zero, errors.Wrapf(err, `判断TRX余额[%s]`, address)
	}
}
func checkBalance(addresses map[string]string) (nonZero map[string]Balance, err error) {
	nonZero = make(map[string]Balance, len(addresses))

	concurrent := 200
	in := make(chan balanceIn, concurrent)
	out := make(chan balanceOut, concurrent)

	wg := &sync.WaitGroup{}
	finish := &sync.WaitGroup{}

	for i := 0; i < concurrent; i++ {
		wg.Add(1)

		if err = checkSingleBalance(in, out, wg); err != nil {
			return nil, err
		}
	}

	finish.Add(1)

	go func() {
		for data := range out {
			if data.error != nil {
				panic(data.error.Error())
			}

			nonZero[data.private] = Balance{
				Address: data.Address,
				USDT:    data.USDT,
				TRX:     data.TRX,
			}
		}

		finish.Done()
	}()

	for private, address := range addresses {
		logger.Info(`处理地址`, zap.String(`地址`, address))

		in <- balanceIn{
			private: private,
			address: address,
		}
	}

	close(in)
	wg.Wait()
	close(out)
	finish.Wait()

	return nonZero, nil
}

type balanceIn struct {
	private string
	address string
}

type balanceOut struct {
	private string
	Balance
	error
}

func loadGrpc() (g *client.GrpcClient, err error) {
	g = client.NewGrpcClient(`35.181.32.79:50051`)

	if err = g.Start(grpc.WithInsecure()); err != nil {
		return nil, errors.Wrap(err, `grpc`)
	}

	return g, nil
}

func checkSingleBalance(in <-chan balanceIn, out chan<- balanceOut, wg *sync.WaitGroup) error {
	g, err := loadGrpc()

	if err != nil {
		return err
	}

	go func() {
		var (
			balances map[string]decimal.Decimal
		)

		for data := range in {
			if balances, err = getBalance(data.address, g); err != nil {
				out <- balanceOut{
					private: data.private,
					error:   err,
				}

				continue
			}

			if balances[USDT].LessThanOrEqual(decimal.Zero) && balances[USDC].LessThanOrEqual(decimal.Zero) && balances[TRX].LessThanOrEqual(decimal.Zero) { //nolint:lll
				continue
			}

			out <- balanceOut{
				private: data.private,
				Balance: Balance{
					Address: data.address,
					USDT:    balances[USDT],
					TRX:     balances[TRX],
					USDC:    balances[USDC],
				},
				error: nil,
			}
		}

		wg.Done()
	}()

	return nil
}

func getBalance(address string, g *client.GrpcClient) (balances map[string]decimal.Decimal, err error) {
	var (
		balance decimal.Decimal
	)

	balances = make(map[string]decimal.Decimal, 1)

	if balance, err = getUSDTBalance(address, USDT, g); err != nil {
		return nil, errors.Wrap(err, `USDT`)
	}

	balances[USDT] = balance

	if balance, err = getUSDTBalance(address, USDC, g); err != nil {
		return nil, errors.Wrap(err, `USDC`)
	}

	balances[USDC] = balance

	if balance, err = getTRXBalance(address, g); err != nil {
		return nil, errors.Wrap(err, `TRX`)
	}

	balances[TRX] = balance

	return balances, nil
}

func load() (addresses map[string]string, err error) {
	var (
		file    *os.File
		records records
	)

	if file, err = os.Open(filePath); err != nil {
		return nil, errors.WithMessage(err, `打开文件`)
	}

	defer func() {
		_ = file.Close()
	}()

	if err = json.NewDecoder(file).Decode(&records); err != nil {
		return nil, errors.Wrap(err, `解析`)
	}

	addresses = make(map[string]string, len(records.Records))

	for _, record := range records.Records {
		addresses[record.PrivateKey] = record.Address
	}

	return addresses, nil
}

type records struct {
	Records []record `json:"RECORDS"`
}

type record struct {
	Id         int    `json:"id"`
	UserID     int    `json:"user_id"`
	Protocol   string `json:"protocol"`
	Mnemonic   string `json:"mnemonic"`
	PrivateKey string `json:"private_key"`
	Address    string `json:"address"`
}

func transfer(path, to string) error {
	var (
		file      *os.File
		balances  = map[string]Balance{}
		err       error
		singleErr error
	)

	if file, err = os.Open(path); err != nil {
		return errors.WithMessage(err, `打开文件`)
	}

	defer func() {
		_ = file.Close()
	}()

	if err = json.NewDecoder(file).Decode(&balances); err != nil {
		return errors.Wrap(err, `解析`)
	}

	concurrent := 100
	in := make(chan transferIn, concurrent)
	out := make(chan error, concurrent)
	wg := &sync.WaitGroup{}

	finish := &sync.WaitGroup{}

	finish.Add(1)

	go func() {
		for singleErr = range out {
			err = multierr.Append(err, singleErr)
		}

		finish.Done()
	}()

	for i := 0; i < concurrent; i++ {
		wg.Add(1)

		if err = transferWorker(in, wg, to, out); err != nil {
			return err
		}
	}

	for private, balance := range balances {
		in <- transferIn{
			address: balance.Address,
			private: private,
		}
	}

	close(in)
	wg.Wait()
	close(out)
	finish.Wait()

	return err
}

type transferIn struct {
	address string
	private string
}

func transferWorker(in <-chan transferIn, wg *sync.WaitGroup, to string, out chan<- error) error {
	g, err := loadGrpc()

	if err != nil {
		return err
	}

	k := keystore.NewKeyStore("keystore", keystore.LightScryptN, keystore.LightScryptP)

	var (
		trx decimal.Decimal
	)

	go func() {
		for data := range in {
			if trx, err = getTRXBalance(data.address, g); err != nil {
				out <- errors.Wrapf(err, `查询[%s]`, data.address)
				continue
			}

			if trx.LessThanOrEqual(decimal.Zero) {
				continue
			}

			if _, err = TransferTrx(data.address, data.private, to, trx, g, k); err != nil {
				out <- errors.Wrapf(err, `转账[%s]`, data.address)
			}
		}

		wg.Done()
	}()

	return nil
}

func TransferTrx(from, privateKeyHex, to string, amount decimal.Decimal, g *client.GrpcClient, k *keystore.KeyStore) (transactionID string, err error) { // nolint:golint,lll
	var (
		account    keystore.Account
		privateKey *ecdsa.PrivateKey
		tx         *api.TransactionExtention
		msg        string
	)

	logger.Info(`转账TRX`, zap.Strings(`from/private/to`, []string{from, privateKeyHex, to}))

	if privateKey, err = crypto.HexToECDSA(privateKeyHex); err != nil {
		return "", errors.Wrapf(err, "解析私钥错误")
	}

	fromAddress := tronAddress.PubkeyToAddress(privateKey.PublicKey)

	if fromAddress.String() != from {
		return "", errors.New("公钥和私钥不匹配")
	}

	if k.HasAddress(fromAddress) {
		account, err = k.Find(keystore.Account{Address: fromAddress})
		msg = `加载账号`
	} else {
		account, err = k.ImportECDSA(privateKey, "")
		msg = `导入账号`
	}

	if err != nil {
		return ``, errors.Wrap(err, msg)
	}

	if err = k.Unlock(account, ``); err != nil {
		return ``, errors.Wrap(err, `解锁`)
	}

	defer func() {
		_ = k.Lock(fromAddress)
	}()

	// 这个接口的数值单位是sun， 1 TRX = 1,000,000 SUN https://tronstation.io/
	if tx, err = g.Transfer(from, to, amount.Mul(decimal.New(1, 6)).IntPart()); err != nil { //nolint:golint,lll
		return "", errors.Wrap(err, "构建交易失败")
	}

	controller := transaction.NewController(g, k, &account, tx.Transaction)

	if err = controller.ExecuteTransaction(); err != nil {
		return "", errors.Wrap(err, `执行交易失败`)
	}

	return strings.TrimPrefix(common.BytesToHexString(tx.GetTxid()), "0x"), nil
}

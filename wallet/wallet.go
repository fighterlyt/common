package wallet

import (
	`context`
	`fmt`
	`gitlab.com/nova_dubai/common/helpers`
	`time`

	`github.com/fighterlyt/gormlogger`
	"github.com/fighterlyt/redislock"
	`github.com/pkg/errors`
	`github.com/shopspring/decimal`
	`gitlab.com/nova_dubai/common/model`
	`gorm.io/gorm`
)

// SyncWalletBalance 同步钱包信息
func SyncWalletBalance(protocol model.Protocol, address string, locker redislock.Locker, currency string, db *gorm.DB, getUserID func(address string) (int64, error), getBalanceFunc func(address, currency string) (decimal.Decimal, error)) error {
	userID, err := getUserID(address)
	if err != nil {
		return errors.Wrap(err, "查询用户ID失败")
	}

	var currencies = []string{currency}
	switch protocol {
	case model.Trc20:
		currencies = append(currencies, model.TRX)
	case model.Erc20:
		currencies = append(currencies, model.ETH)
	default:
		return fmt.Errorf("不支持的协议[%s]", protocol)
	}

	// 查币种余额
	for i := range currencies {
		if err = checkAndSaveSingleSymbolBalance(protocol, address, currencies[i], userID, locker, db, getBalanceFunc); err != nil {
			return errors.Wrap(err, "查询账户余额失败")
		}
	}

	return nil
}

// 查询单个币种的余额
func checkAndSaveSingleSymbolBalance(protocol model.Protocol, address string, currency string, userID int64, locker redislock.Locker, db *gorm.DB, getBalanceFunc func(address, currency string) (decimal.Decimal, error)) error {
	mutex, err := locker.GetMutex(fmt.Sprintf("balance_%s_%d_%s", protocol, userID, currency), 2*time.Second)
	if err != nil {
		return errors.Wrap(err, "加锁失败")
	}

	// 加锁
	helpers.EnsureRedisLock(mutex)

	// 解锁
	defer mutex.UnLock()

	var balance decimal.Decimal

	// 查余额
	balance, err = getBalanceFunc(address, currency)

	if err != nil {
		return errors.Wrap(err, "更新余额失败")
	}

	// 更新余额
	var (
		userBalance UserBalance
		ctx         = context.WithValue(context.Background(), gormlogger.IgnoreErrorKey, gorm.ErrRecordNotFound)
	)
	if err = db.WithContext(ctx).Where("userID=?", userID).Where("protocol=?", protocol).Where("symbol=?", currency).First(&userBalance).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.Wrap(err, "查询用户余额失败")
		}

		userBalance = UserBalance{
			UserID:   userID,
			Protocol: protocol.String(),
			Symbol:   currency,
			Balance:  balance,
		}
	}

	userBalance.Balance = balance

	if err = db.Save(&userBalance).Error; err != nil {
		return errors.Wrap(err, "保存用户余额信息失败")
	}

	return nil
}

package summaryextend

import (
	"fmt"

	"github.com/fighterlyt/common/helpers"
	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/youthlin/t"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// client 客户端
type client struct {
	tableName string     // 表名
	slot      Slot       // 槽位类型
	model     Detail     // model
	logger    log.Logger // 日志器
	db        *gorm.DB   // db
	bigInt    bool
}

func (m *client) SetBigInt(on bool) {
	m.bigInt = on
}

func (m client) newData() Summary {
	if m.bigInt {
		return &DetailWithBigInt{}
	}

	return &Detail{}
}

func (m client) GetSummaryByDate(date int, ownerID int64) (records Summary, err error) {
	data := m.newData()

	if err = m.db.Session(&gorm.Session{}).Where("ownerID = ? and slotValue = ?", ownerID, date).First(data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrap(err, "查询用户汇总失败")
	}

	return Summary(data), nil
}

func (m client) GetAllOrderedOwnerID(date int) (allOwnerID []int64, err error) {
	// 这里的排序很重要
	if err = m.db.Session(&gorm.Session{}).Where("slotValue = ?", date).Select("ownerID").Group("ownerID").Order("ownerID").Find(&allOwnerID).Error; err != nil { //nolint:lll
		return nil, errors.Wrap(err, "获取所有用户失败")
	}

	return allOwnerID, nil
}

func (m client) GetAllSlotValue() (allSlotValue []int, err error) {
	if err = m.db.Session(&gorm.Session{}).Select("slotValue").Group("slotValue").Find(&allSlotValue).Error; err != nil {
		return nil, errors.Wrap(err, "获取所有date失败")
	}

	return allSlotValue, nil
}

func NewClient(tableName string, slot Slot, logger log.Logger, db *gorm.DB) (result *client, err error) {
	if db == nil {
		return nil, errors.New(`db不能为空`)
	}

	if logger == nil {
		return nil, errors.New(`日志器不能为空`)
	}

	if tableName == `` {
		return nil, errors.New(`表名不能为空`)
	}

	model := Detail{
		tableName: tableName,
	}

	db = db.Table(tableName)

	if err = db.AutoMigrate(model); err != nil {
		return nil, errors.Wrap(err, `创建表`)
	}

	return &client{
		tableName: tableName,
		slot:      slot,
		logger:    logger,
		db:        db,
		model:     model,
	}, nil
}

/*
Summarize 汇总
参数:
*	ownerID	string          所有者
*	amount 	decimal.Decimal	值
返回值:
*	error  	error          	错误
*/
func (m *client) Summarize(ownerID string, amount decimal.Decimal, extendValue ...decimal.Decimal) error {
	return m.summarize(ownerID, amount, 1, extendValue...)
}

func (m *client) SummarizeOptimism(ownerID string, amount decimal.Decimal, extendValue ...decimal.Decimal) error {
	return m.summarizeFirstUpdate(ownerID, amount, 1, extendValue...)
}

func (m *client) summarizeFirstUpdate(ownerID string, amount decimal.Decimal, times int, extendValue ...decimal.Decimal) error {
	slotValue, err := m.getSlotValue(ownerID)
	if err != nil {
		return errors.Wrap(err, `计算slotValue错误`)
	}

	data, updates, err := m.buildSummarizeDayParams(slotValue, ownerID, amount, times, extendValue...)
	if err != nil {
		return errors.Wrap(err, "构建汇总数据失败")
	}

	db := m.db.Session(&gorm.Session{}).Where("ownerID=? and slotValue=?", ownerID, slotValue).Updates(updates)

	if err = db.Error; err != nil {
		return errors.Wrap(err, "更新失败")
	}

	if db.RowsAffected > 0 { // 更新到，则返回
		return nil
	}

	return m.db.Session(&gorm.Session{}).Create(data).Error
}

func (m *client) SummarizeNotAddTimes(ownerID string, amount decimal.Decimal, extendValue ...decimal.Decimal) error {
	return m.summarize(ownerID, amount, 0, extendValue...)
}

func (m *client) RevertSummarize(ownerID string, amount decimal.Decimal, extendValue ...decimal.Decimal) error {
	var extendValue2 []decimal.Decimal
	for _, d := range extendValue {
		extendValue2 = append(extendValue2, d.Neg())
	}

	return m.summarize(ownerID, amount.Neg(), -1, extendValue2...)
}

func (m *client) RevertSummarizeOptimism(ownerID string, amount decimal.Decimal, extendValue ...decimal.Decimal) error {
	var extendValue2 []decimal.Decimal
	for _, d := range extendValue {
		extendValue2 = append(extendValue2, d.Neg())
	}

	return m.summarizeFirstUpdate(ownerID, amount.Neg(), -1, extendValue2...)
}

func (m *client) summarize(ownerID string, amount decimal.Decimal, times int, extendValue ...decimal.Decimal) error {
	var (
		slotValue string
		err       error
	)

	if slotValue, err = m.getSlotValue(ownerID); err != nil {
		return errors.Wrap(err, `计算slotValue错误`)
	}

	data := newSummary(m.slot, ownerID, amount, slotValue)

	db := m.db.Session(&gorm.Session{})

	updates := map[string]interface{}{
		"value": gorm.Expr(`value + ?`, amount),
	}

	if times != 0 {
		updates["times"] = gorm.Expr(`times + ?`, times)
	}

	for i, extend := range extendValue {
		key := fmt.Sprintf(`value_%d`, i+1)

		if extend.IsZero() {
			continue
		}

		updates[key] = gorm.Expr(fmt.Sprintf(`%s + ?`, key), extend)

		if err = data.SetExtendValue(i, extend); err != nil {
			return err
		}
	}

	// 写入或者更新
	err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: `ownerID`}, {Name: `slotValue`}},
		DoUpdates: clause.Assignments(updates),
	}).Create(&data).Error
	if err != nil {
		return errors.WithMessage(err, m.tableName)
	}

	return nil
}

func (m *client) RevertSummarizeDay(date int64, ownerID string, amount decimal.Decimal, extendValue ...decimal.Decimal) error {
	var extendValue2 []decimal.Decimal
	for _, d := range extendValue {
		extendValue2 = append(extendValue2, d.Neg())
	}

	return m.summarizeDay(date, ownerID, amount.Neg(), -1, extendValue2...)
}

func (m *client) buildSummarizeDayParams(slotValue, ownerID string, amount decimal.Decimal, times int, extendValue ...decimal.Decimal) (detail *Detail, updates map[string]interface{}, err error) { //nolint:lll
	detail = newSummary(m.slot, ownerID, amount, slotValue)

	updates = map[string]interface{}{
		"value": gorm.Expr(`value + ?`, amount),
	}

	if times != 0 {
		updates["times"] = gorm.Expr(`times + ?`, times)
	}

	for i, extend := range extendValue {
		key := fmt.Sprintf(`value_%d`, i+1)

		if extend.IsZero() {
			continue
		}

		updates[key] = gorm.Expr(fmt.Sprintf(`%s + ?`, key), extend)

		if err = detail.SetExtendValue(i, extend); err != nil {
			return nil, nil, err
		}
	}

	return detail, updates, nil
}

func (m *client) summarizeDay(date int64, ownerID string, amount decimal.Decimal, times int, extendValue ...decimal.Decimal) error {
	data, updates, err := m.buildSummarizeDayParams(fmt.Sprintf(`%d`, date), ownerID, amount, times, extendValue...)
	if err != nil {
		return errors.Wrap(err, t.T("构建汇总数据失败"))
	}

	db := m.db.Session(&gorm.Session{})

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: `ownerID`}, {Name: `slotValue`}},
		DoUpdates: clause.Assignments(updates),
	}).Create(&data).Error
}

func (m client) SummarizeDayOptimism(date int64, ownerID string, amount decimal.Decimal, extendValue ...decimal.Decimal) error {
	data, updates, err := m.buildSummarizeDayParams(fmt.Sprintf(`%d`, date), ownerID, amount, 1, extendValue...)
	if err != nil {
		return errors.Wrap(err, "构建汇总数据失败")
	}

	db := m.db.Session(&gorm.Session{}).Where("ownerID=? and slotValue=?", ownerID, date).Updates(updates)

	if err = db.Error; err != nil {
		return errors.Wrap(err, "更新失败")
	}

	if db.RowsAffected > 0 { // 更新到，则返回
		return nil
	}

	return m.db.Session(&gorm.Session{}).Create(data).Error
}

func (m *client) SummarizeDay(date int64, ownerID string, amount decimal.Decimal, extendValue ...decimal.Decimal) error {
	return m.summarizeDay(date, ownerID, amount, 1, extendValue...)
}

func (m client) getSlotValue(userID string) (value string, err error) {
	switch m.slot {
	case SlotDay: // 如果是天,那么就是yyyy-mm-dd
		return fmt.Sprintf(`%d`, helpers.GetDateInDefault()), nil
	case SlotWhole:
		return userID, nil
	case SlotMonth:
		return fmt.Sprintf(`%d`, helpers.GetMonthInDefault()), nil
	default:
		return ``, newErrNotSupportSlot(m.slot)
	}
}

func (m client) Key() string {
	return m.tableName
}

func (m client) Model() Summary {
	return &m.model
}

func (m client) GetSummary(ownerIDs []string, from, to int64) (records []Summary, err error) {
	query := m.db.Session(&gorm.Session{})
	if len(ownerIDs) > 0 {
		query = query.Where(`ownerID in ?`, ownerIDs)
	}

	var (
		data []*Detail
	)

	if query, err = m.buildScopeByRange(from, to, query); err != nil {
		return nil, errors.Wrap(err, `构建时间查询`)
	}

	if err = query.Find(&data).Error; err != nil {
		return nil, errors.Wrap(err, `数据库操作`)
	}

	records = make([]Summary, 0, len(data))

	for i := range data {
		records = append(records, data[i])
	}

	return records, nil
}

func (m client) GetSummaryByLike(like string, from, to int64) (records []Summary, err error) {
	query := m.db.Session(&gorm.Session{}).Where(`ownerID like ?`, "%"+like+"%")

	var (
		data []*Detail
	)

	if query, err = m.buildScopeByRange(from, to, query); err != nil {
		return nil, errors.Wrap(err, `构建时间查询`)
	}

	if err = query.Find(&data).Error; err != nil {
		return nil, errors.Wrap(err, `数据库操作`)
	}

	records = make([]Summary, 0, len(data))

	for i := range data {
		records = append(records, data[i])
	}

	return records, nil
}

func (m client) buildScopeByRange(from, to int64, db *gorm.DB) (query *gorm.DB, err error) {
	var (
		scope helpers.Scope
	)

	if from == 0 && to == 0 { // 没有时间戳，全部
		return db, nil
	}

	if from != 0 && to == 0 { // 有开始无截止
		return db.Where(`slotValue >= ?`, helpers.GetDateByTime(from)), nil
	}

	if from == 0 && to != 0 { // 有截止无开始
		return db.Where(`slotValue <= ?`, helpers.GetDateByTime(to)), nil
	}

	if scope, err = m.getSlotValueByRange(from, to); err != nil { // 有开始有截止
		return nil, errors.Wrapf(err, `使用[%s][%d]-[%d]计算slotValue错误`, m.slot, from, to)
	}

	if scope != nil {
		return db.Scopes(scope), nil
	}

	return db, nil
}
func (m client) getSlotValueByRange(from, to int64) (scope helpers.Scope, err error) {
	switch m.slot {
	case SlotDay:
		if from >= to {
			return nil, fmt.Errorf(`开始时间[%d]必须小于结束时间[%d]`, from, to)
		}

		ranges := helpers.GetDatesByRange(from, to)

		result := make([]string, 0, len(ranges))

		for _, item := range ranges {
			result = append(result, fmt.Sprintf(`%d`, item))
		}

		return func(db *gorm.DB) *gorm.DB {
			return db.Where(`slotValue in (?)`, result)
		}, nil
	case SlotMonth:
		if from >= to {
			return nil, fmt.Errorf(`开始时间[%d]必须小于结束时间[%d]`, from, to)
		}

		return func(db *gorm.DB) *gorm.DB {
			return db.Where(`slotValue >= ? and slotValue <= ?`, fmt.Sprintf(`%d`, helpers.GetMonthByTime(from)), fmt.Sprintf(`%d`, helpers.GetMonthByTime(to))) //nolint:lll
		}, nil
	case SlotWhole:
		return nil, nil
	default:
		return nil, newErrNotSupportSlot(m.slot)
	}
}

type ErrNotSupportSlot struct {
	slot Slot
}

func newErrNotSupportSlot(slot Slot) *ErrNotSupportSlot {
	return &ErrNotSupportSlot{slot: slot}
}

func (e ErrNotSupportSlot) Error() string {
	return fmt.Sprintf(`不支持的Slot[%s]`, e.slot)
}

func (m client) GetSummarySummary(ownerIDs []string, from, to int64) (record Summary, err error) {
	query := m.db.Session(&gorm.Session{})
	if len(ownerIDs) > 0 {
		query = query.Where(`ownerID in ?`, ownerIDs)
	}

	data := m.newData()

	if query, err = m.buildScopeByRange(from, to, query); err != nil {
		return nil, errors.Wrap(err, `构建时间查询`)
	}

	if err = query.Select(
		`sum(value) as value,` +
			`sum(value_1) as value_1,` +
			`sum(value_2) as value_2,` +
			`sum(value_3) as value_3,` +
			`sum(value_4) as value_4,` +
			`sum(value_5) as value_5,` +
			`sum(value_6) as value_6,` +
			`sum(value_7) as value_7,` +
			`sum(value_8) as value_8,` +
			`sum(value_9) as value_9,` +
			`sum(value_10) as value_10,` +
			`sum(times) as times`).Find(data).Error; err != nil {
		return nil, errors.Wrap(err, `数据库操作`)
	}

	return data, nil
}

func (m client) GetSummaryExclude(excludeOwnerID []string, from, to int64, selects ...string) (records []Summary, err error) {
	query := m.db.Session(&gorm.Session{})
	if len(excludeOwnerID) > 0 {
		query = query.Where(`ownerID not in ?`, excludeOwnerID)
	}

	var (
		data       []*Detail
		bigIntData []*DetailWithBigInt
		length     int
	)

	if query, err = m.buildScopeByRange(from, to, query); err != nil {
		return nil, errors.Wrap(err, `构建时间查询`)
	}

	for _, item := range selects {
		query = query.Select(item)
	}

	if m.bigInt {
		err = query.Find(&bigIntData).Error
		length = len(bigIntData)
	} else {
		err = query.Find(&data).Error
		length = len(data)
	}

	if err != nil {
		return nil, errors.Wrap(err, `数据库操作`)
	}

	records = make([]Summary, 0, length)

	if m.bigInt {
		for i := range bigIntData {
			records = append(records, bigIntData[i])
		}
	} else {
		for i := range data {
			records = append(records, data[i])
		}
	}

	return records, nil
}

package helpers

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	beijin          *time.Location
	beijingLocation = `Asia/Shanghai`
	loadBeiJinLock  = &sync.Mutex{}
)

func NowInBeiJin() time.Time {
	if beijin == nil {
		loadBeiJinLock.Lock()
		defer loadBeiJinLock.Unlock()

		if beijin == nil {
			var err error

			if beijin, err = time.LoadLocation(beijingLocation); err != nil {
				panic(fmt.Sprintf(`time.LoadLocation(%s) [%s]`, beijingLocation, err.Error()))
			}
		}
	}

	return time.Now().In(beijin)
}

/*GetBeiJin 获取北京时区
参数:
返回值:
*	*time.Location	*time.Location	返回值1
*/
func GetBeiJin() *time.Location {
	if beijin == nil {
		loadBeiJinLock.Lock()
		defer loadBeiJinLock.Unlock()

		if beijin == nil {
			var err error

			if beijin, err = time.LoadLocation(beijingLocation); err != nil {
				panic(fmt.Sprintf(`time.LoadLocation(%s) [%s]`, beijingLocation, err.Error()))
			}
		}
	}

	return beijin
}

/*GetStartOfDayInLocation 获取指定时区的指定天的第一秒
参数:
*	date    	int64         	参数1
*	location	*time.Location	参数2
返回值:
*	start   	time.Time     	返回值1
*	err     	error         	返回值2
*/
func GetStartOfDayInLocation(date int, location *time.Location) (start time.Time, err error) {
	var (
		day time.Time
	)

	if day, err = time.Parse(`20060102`, fmt.Sprintf(`%d`, date)); err != nil {
		return start, errors.Wrapf(err, `解析天[%d]`, date)
	}

	start = time.Unix(day.Unix()/(secondsInDay)*secondsInDay, 0)
	start = start.Add(time.Duration(int64(time.Second) * -1 * int64(offset(location))))

	return start, nil
}

func offset(location *time.Location) int {
	_, offset := time.Now().In(location).Zone()
	return offset
}

const (
	secondsInDay = 24 * 60 * 60
	layout       = "2006-01-02 15:04:05"
)

func GetDate() int {
	now := NowInBeiJin()
	return now.Year()*10000 + int(now.Month())*100 + now.Day()
}

// FormatDateTime 转换时间格式 date格式为yyyymmdd
func FormatDateTime(date int) time.Time {
	return time.Date(date/10000, time.Month(date%10000/100), date%100, 0, 0, 0, 0, GetBeiJin())
}

func GetDateByTime(t int64) int {
	now := time.Unix(t, 0).In(GetBeiJin())

	return now.Year()*10000 + int(now.Month())*100 + now.Day()
}

/*GetDatesByRange 通过时间戳获取中间的日期
参数:
*	from 	int64	开始时间戳,单位秒,结果包括start所在的天
*	to   	int64	结束时间戳,单位秒,结果包括to所在的天
返回值:
*	[]int	[]int	天，按照增序排列
*/
func GetDatesByRange(from, to int64) []int {
	if from >= to {
		return nil
	}

	fromDate := GetDateByTime(from)
	toDate := GetDateByTime(to)

	result := make([]int, 0, 10)

	for ; fromDate <= toDate; fromDate = DataCal(fromDate, 1) {
		result = append(result, fromDate)
	}

	return result
}

type Time int64

func (t Time) IsZero() bool {
	return t <= 0
}

func Now() Time {
	return Time(time.Now().Unix())
}
func (t Time) MarshalText() ([]byte, error) {
	if t == 0 {
		return nil, nil
	}

	return []byte(time.Unix(int64(t), 0).In(beijin).Format(`2006-01-02 15:04:05`)), nil
}

func (t Time) Value() (driver.Value, error) {
	return driver.Int32.ConvertValue(int32(t))
}

func (t Time) Unix() int64 {
	return int64(t)
}

func DataCal(date, add int) int {
	now := time.Date(date/10000, time.Month(date%10000/100), date%100, 0, 0, 0, 0, GetBeiJin())

	now = now.AddDate(0, 0, add)

	return now.Year()*10000 + int(now.Month())*100 + now.Day()
}

func TimeToBeijingDate(t time.Time) int {
	beijingt := t.In(beijin)
	return beijingt.Year()*10000 + int(beijingt.Month())*100 + beijingt.Day()
}

func (t *Time) UnmarshalJSON(data []byte) error {
	data = bytes.Trim(data, `""`)
	s := string(data)
	if len(data) == 8 {
		_, err := time.ParseInLocation("20060102", s, beijin)
		if err != nil {
			return err
		}
		parseInt, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*t = Time(parseInt)
		return nil
	}
	if len(data) == 10 {
		parseInt, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*t = Time(parseInt)
		return nil
	}
	value, err := time.ParseInLocation(layout[:len(data)], s, beijin)

	if err != nil {
		return err
	}
	*t = Time(value.Unix())

	return nil
}

func (t *Time) Scan(src interface{}) error {
	finished := false
	v := reflect.ValueOf(src)

	switch v.Kind() {

	case reflect.Int64, reflect.Int32:
		*t = Time(v.Int())
		finished = true

	case reflect.Slice:
		data, ok := src.([]uint8)
		if ok {
			value := int64(0)

			for i := range data {
				// 首先是16进制，然后是ASCII码(31对应1)
				value = value*10 + int64(data[i]/16*10+data[i]%16-30)
			}

			*t = Time(value)
			finished = true
		}
	}

	if finished {
		return nil
	}

	return fmt.Errorf(`参数类型不支持,实际是[%s]`, reflect.TypeOf(src).Kind().String())
}

func BeginningOfMonday(now time.Time) time.Time {
	weekday := int(now.Weekday())

	if weekday == 0 {
		weekday = 7
	}

	monday := now.AddDate(0, 0, -weekday+1)

	return monday
}

func EndOfWeek(now time.Time) time.Time {
	return BeginningOfMonday(now).AddDate(0, 0, 6)
}

func LastWeekBeginningOfMonday(now time.Time) time.Time {
	return BeginningOfMonday(now).AddDate(0, 0, -7)
}

func EndOfLastWeek(now time.Time) time.Time {
	return DecreaseOneDay(BeginningOfMonday(now))
}

func BeginningOfMonth(now time.Time) time.Time {
	y, m, _ := now.Date()
	month1st := time.Date(y, m, 1, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())

	return month1st
}

func EndOfMonth(now time.Time) time.Time {
	return DecreaseOneDay(BeginningOfMonth(now).AddDate(0, 1, 0))
}

func LastMonthBeginningOfFirst(now time.Time) time.Time {
	y, m, _ := now.Date()
	month1st := time.Date(y, m-1, 1, 0, 0, 0, 0, now.Location())

	return month1st
}

func EndOfLastMonth(now time.Time) time.Time {
	return DecreaseOneDay(BeginningOfMonth(now))
}

func BeginningOfDay() time.Time {
	now := NowInBeiJin()
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, now.Location())
}

func EndOfDay(now time.Time) time.Time {
	y, m, d := now.Date()
	return time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), now.Location())
}

func DecreaseOneDay(now time.Time) time.Time {
	return now.AddDate(0, 0, -1)
}

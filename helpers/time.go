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

/*NowInBeiJin 北京时间同时设置北京时区
参数:
返回值:
*	time.Time	time.Time	time
*/
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
*	*time.Location	*time.Location	北京时区
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
*	date    	int64         	int类型的date
*	location	*time.Location	时区
返回值:
*	start   	time.Time     	指定时区的指定天的第一秒time
*	err     	error         	错误
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

/*offset 在指定时区便宜了多少秒
参数:
*	location	*time.Location	时区
返回值:
*	int     	int           	偏移量(秒)
*/

func offset(location *time.Location) int {
	_, offset := time.Now().In(location).Zone()
	return offset
}

const (
	secondsInDay = 24 * 60 * 60
	layout       = "2006-01-02 15:04:05"
)

/*GetDate 获取北京时间int类型的日期
参数:
返回值:
*	int	int	int类型的日期
*/
func GetDate() int {
	now := NowInBeiJin()
	return now.Year()*before4Mask + int(now.Month())*after2Mask + now.Day()
}

// FormatDateTime 转换时间格式 date格式为yyyymmdd
func FormatDateTime(date int) time.Time {
	return time.Date(date/before4Mask, time.Month(date%before4Mask/after2Mask), date%after2Mask, 0, 0, 0, 0, GetBeiJin())
}

/*GetDateByTime 获取时间戳在北京时区的int类型日期
参数:
*	t  	int64	时间戳
返回值:
*	int	int  	int类型的日期
*/
func GetDateByTime(t int64) int {
	now := time.Unix(t, 0).In(GetBeiJin())

	return now.Year()*before4Mask + int(now.Month())*after2Mask + now.Day()
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

/*IsZero 是否为0值
参数:
返回值:
*	bool	bool 是否为0值
*/
func (t Time) IsZero() bool {
	return t <= 0
}

/*Now 获取当前时间戳
参数:
返回值:
*	Time	Time	当前时间戳
*/
func Now() Time {
	return Time(time.Now().Unix())
}

/*MarshalText 序列化方法
参数:
返回值:
*	[]byte	[]byte	序列化之后的值
*	error 	error 	错误
*/
func (t Time) MarshalText() ([]byte, error) {
	if t == 0 {
		return nil, nil
	}

	return []byte(time.Unix(int64(t), 0).In(GetBeiJin()).Format(`2006-01-02 15:04:05`)), nil
}

/*Value 放入数据库值的序列化方法
参数:
返回值:
*	driver.Value	driver.Value	数据库值类型
*	error       	error       	错误
*/
func (t Time) Value() (driver.Value, error) {
	return driver.Int32.ConvertValue(int32(t))
}

/*Unix 时间戳
参数:
返回值:
*	int64	int64	时间戳
*/
func (t Time) Unix() int64 {
	return int64(t)
}

const (
	before4Mask = 10000 // 前4位
	after2Mask  = 100   // 后2位
)

/*DataCal 日期增加或者减少天
参数:
*	date	int int类型的日期
*	add 	int 增加或者减少的天数
返回值:
*	int 	int 增加或减少后的int类型日期
*/
func DataCal(date, add int) int {
	now := time.Date(date/before4Mask, time.Month(date%before4Mask/after2Mask), date%after2Mask, 0, 0, 0, 0, GetBeiJin())

	now = now.AddDate(0, 0, add)

	return now.Year()*before4Mask + int(now.Month())*after2Mask + now.Day()
}

/*TimeToBeijingDate time类型转换成北京时区int日期
参数:
*	t  	time.Time	time
返回值:
*	int	int      	int类型的Date
*/
func TimeToBeijingDate(t time.Time) int {
	beijingt := t.In(beijin)
	return beijingt.Year()*before4Mask + int(beijingt.Month())*after2Mask + beijingt.Day()
}

const dateStrLen = 8

/*UnmarshalJSON 反序列化方法 支持int类型的日期 时间戳 2006-01-02 15:04:05 类型的字符串
参数:
*	data 	[]byte	入参
返回值:
*	error	error 	错误
*/
func (t *Time) UnmarshalJSON(data []byte) error {
	data = bytes.Trim(data, `""`)
	s := string(data)

	if len(data) == dateStrLen {
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

/*Scan 数据库类型变成go类型方法
参数:
*	src  	interface{}	数据库类型
返回值:
*	error	error      	错误
*/
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
	default:
		return errors.Errorf("un support type %T", v.Kind())
	}

	if finished {
		return nil
	}

	return fmt.Errorf(`参数类型不支持,实际是[%s]`, reflect.TypeOf(src).Kind().String())
}

/*BeginningOfMonday 传入时间本周第一天当前时间
参数:
*	now      	time.Time	当前时间
返回值:
*	time.Time	time.Time	时间本周第一天的当前时间
*/

func BeginningOfMonday(now time.Time) time.Time {
	weekday := int(now.Weekday())

	if weekday == 0 {
		weekday = perWeek
	}

	monday := now.AddDate(0, 0, -weekday+1)

	return monday
}

const (
	endOfWeekFromMonday = 6
	perWeek             = 7
)

/*EndOfWeek 获取本周日结束时间
参数:
*	now      	time.Time	当前时间
返回值:
*	time.Time	time.Time	本周日的当前时间
*/
func EndOfWeek(now time.Time) time.Time {
	return BeginningOfMonday(now).AddDate(0, 0, endOfWeekFromMonday)
}

/*LastWeekBeginningOfMonday 获取上周的星期一当前时间
参数:
*	now      	time.Time	当前时间
返回值:
*	time.Time	time.Time	上周的星期一的当前时间
*/
func LastWeekBeginningOfMonday(now time.Time) time.Time {
	return BeginningOfMonday(now).AddDate(0, 0, -perWeek)
}

/*EndOfLastWeek 获取上周日的当前时间
参数:
*	now      	time.Time	当前时间
返回值:
*	time.Time	time.Time	上周日的当前时间
*/
func EndOfLastWeek(now time.Time) time.Time {
	return DecreaseOneDay(BeginningOfMonday(now))
}

/*BeginningOfMonth 获取本月初的当前时间
参数:
*	now      	time.Time	当前时间
返回值:
*	time.Time	time.Time	本月初的当前时间
*/
func BeginningOfMonth(now time.Time) time.Time {
	y, m, _ := now.Date()
	month1st := time.Date(y, m, 1, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())

	return month1st
}

/*EndOfMonth 获取本月底的当前时间
参数:
*	now      	time.Time	当前时间
返回值:
*	time.Time	time.Time	本月底的当前时间
*/
func EndOfMonth(now time.Time) time.Time {
	return DecreaseOneDay(BeginningOfMonth(now).AddDate(0, 1, 0))
}

/*LastMonthBeginningOfFirst 获取上月初的0时
参数:
*	now      	time.Time	当前时间
返回值:
*	time.Time	time.Time	上月初的0时
*/
func LastMonthBeginningOfFirst(now time.Time) time.Time {
	y, m, _ := now.Date()
	month1st := time.Date(y, m-1, 1, 0, 0, 0, 0, now.Location())

	return month1st
}

/*EndOfLastMonth 获取上月底的当前时间
参数:
*	now      	time.Time	当前时间
返回值:
*	time.Time	time.Time	上月底的当前时间
*/
func EndOfLastMonth(now time.Time) time.Time {
	return DecreaseOneDay(BeginningOfMonth(now))
}

/*BeginningOfDay 获取今天的0时
参数:
返回值:
*	time.Time	time.Time	当前时间
*/
func BeginningOfDay() time.Time {
	now := NowInBeiJin()
	y, m, d := now.Date()

	return time.Date(y, m, d, 0, 0, 0, 0, now.Location())
}

const (
	endOfDayHour = 23
	endOfDayMin  = 59
	endOfDaySec  = 59
)

/*EndOfDay 获取今天的23:59:59
参数:
*	now      	time.Time	当前时间
返回值:
*	time.Time	time.Time	今天的23:59:59
*/
func EndOfDay(now time.Time) time.Time {
	y, m, d := now.Date()

	return time.Date(y, m, d, endOfDayHour, endOfDayMin, endOfDaySec, int(time.Second-time.Nanosecond), now.Location())
}

/*DecreaseOneDay 减少1天
参数:
*	now      	time.Time	当前时间
返回值:
*	time.Time	time.Time	减少一天
*/
func DecreaseOneDay(now time.Time) time.Time {
	return now.AddDate(0, 0, -1)
}

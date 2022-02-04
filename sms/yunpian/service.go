package yunpian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	stderror "errors"

	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.com/nova_dubai/common/helpers"
	"gitlab.com/nova_dubai/common/sms"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

const (
	sendSMSURL    = `https://us.yunpian.com/v2/sms/single_send.json`
	balanceURL    = `https://us.yunpian.com/v2/user/get.json`
	pullStatusURL = `https://us.yunpian.com/v2/sms/pull_status.json`
	pageSize      = `50`
)

var (
	bg    = context.Background()
	debug = false
)

// Service 短信服务
type Service struct {
	apiKey             string           // api Key
	client             *http.Client     // http 客户端
	timeout            time.Duration    // 超时
	logger             log.Logger       // 日志器
	recordService      sms.RecordAccess // 记录更新
	pullStatusInterval time.Duration    // 拉取状态时的间隔
	retryCheckTimes    int              // 重试查单次数
}

/*NewService 新建服务
参数:
*	apiKey  	        string       	    api Key
*	timeout 	        time.Duration	    超时
*   logger              log.Logger          日志器
*	pullStatusInterval	time.Duration   	拉取状态间隔
*	recordService     	sms.RecordAccess	状态更新
*	retryCheckTimes   	int             	重试次数
返回值:
*	*Service	        *Service     	服务
*/
func NewService(apiKey string, timeout time.Duration, logger log.Logger, pullStatusInterval time.Duration, recordService sms.RecordAccess, retryCheckTimes int) *Service { //nolint:lll
	service := &Service{
		apiKey:             apiKey,
		client:             &http.Client{},
		timeout:            timeout,
		logger:             logger,
		pullStatusInterval: pullStatusInterval,
		recordService:      recordService,
		retryCheckTimes:    retryCheckTimes,
	}

	service.getReport()

	return service
}

func (s Service) DirectSend(_, _ string) error {
	return sms.ErrNotSupported
}

func (s Service) TemplateSend(target, content, id string) error {
	target = strings.ReplaceAll(target, `-`, ``)

	if !strings.HasPrefix(target, `+`) {
		target = `+` + target
	}

	values := url.Values{}
	values.Set(`apikey`, s.apiKey)
	values.Set(`mobile`, target)
	values.Set(`text`, content)
	values.Set(`uid`, id)

	result := &sendResponse{}

	exceeded, err := s.send(sendSMSURL, values, result, debug)

	if err == nil || !exceeded {
		return err
	}

	if s.recordService == nil {
		return errors.New(`fail`)
	}

	s.logger.Info(`超时错误，查询记录`, helpers.ZapError(err))

	var (
		success bool
	)

	if success, err = s.isSendSuccess(id); err != nil {
		return err
	}

	if success {
		return nil
	}

	return errors.New(`fail`)
}

func (s Service) isSendSuccess(id string) (success bool, err error) {
	var (
		finishStatus sms.SendStatus
	)
	// 超时，特殊处理

	for i := 0; i < s.retryCheckTimes; i++ {
		if finishStatus, err = s.recordService.GetFinishStatus(id); err != nil {
			return false, errors.Wrap(err, `query`)
		}

		switch finishStatus {
		case sms.SendFail:
			return false, nil
		case sms.SendSuccess:
			return true, nil
		default:
			time.Sleep(time.Second)
		}
	}

	return false, nil
}

func (s Service) Support(supported sms.Supported) bool {
	switch supported {
	case sms.SupportDirectSend:
		return false
	case sms.SupportTemplateSend:
		return true
	default:
		return false
	}
}

func (s Service) Balance() (balance decimal.Decimal, err error) {
	values := url.Values{}
	values.Set(`apikey`, s.apiKey)

	result := &getResponse{}

	if _, err = s.send(balanceURL, values, result, debug); err != nil {
		return decimal.Zero, err
	}

	return result.Balance, nil
}

func (s Service) send(url string, data url.Values, result result, debug bool) (isExceed bool, err error) {
	ctx, cancel := context.WithTimeout(bg, s.timeout)
	defer cancel()

	var (
		req  *http.Request
		resp *http.Response
	)

	defer func() {
		if err != nil {
			isExceed = stderror.Is(err, context.DeadlineExceeded)
		}
	}()

	if req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(data.Encode())); err != nil {
		return false, err
	}

	req.Header.Set(`contentType`, `application/x-www-form-urlencoded`)

	if resp, err = s.client.Do(req); err != nil {
		return false, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	var (
		response []byte
		reader   io.Reader
	)

	reader = resp.Body

	if debug {
		if response, err = ioutil.ReadAll(resp.Body); err != nil {
			return false, errors.Wrap(err, `ReadAll`)
		}

		s.logger.Info(`应答`, zap.ByteString(`应答`, response))

		reader = bytes.NewReader(response)
	}

	if err = json.NewDecoder(reader).Decode(result); err != nil {
		return false, err
	}

	s.logger.Info(`请求完成`, zap.Any(`结果`, result))

	return false, result.Validate()
}

type sendResponse struct {
	Code   int     `json:"code"`
	Msg    string  `json:"msg"`
	Count  int     `json:"count"`
	Fee    float64 `json:"fee"`
	Unit   string  `json:"unit"`
	Mobile string  `json:"mobile"`
	Sid    int64   `json:"sid"`
}

/*Validate 校验 错误码 https://www.yunpian.com/official/document/sms/zh_CN/returnvalue_common
参数:
返回值:
*	error	error	返回值1
*/
func (s sendResponse) Validate() error {
	if s.Code != 0 {
		return fmt.Errorf(`%d`, s.Code)
	}

	return nil
}

type getResponse struct {
	Nick             string          `json:"nick"`
	GmtCreated       string          `json:"gmt_created"`
	Mobile           string          `json:"mobile"`
	Email            string          `json:"email"`
	Balance          decimal.Decimal `json:"balance"`
	AlarmBalance     int             `json:"alarm_balance"`
	EmergencyContact string          `json:"emergency_contact"`
	EmergencyMobile  string          `json:"emergency_mobile"`
}

func (g getResponse) Validate() error {
	return nil
}

type result interface {
	Validate() error
}

func (s *Service) getReport() {
	if s.recordService == nil {
		return
	}

	go func() {
		defer func() {
			x := recover()
			if x != nil {
				s.logger.Error(`获取发送状态panic`, zap.Any(`值`, x))
				s.getReport()
			}
		}()

		for {
			if err := s.getNewReport(); err != nil {
				s.logger.Error(`获取最新状态错误`, zap.String(`错误`, err.Error()))
			}
		}
	}()
}

func (s Service) getNewReport() error {
	result, err := s.pullStatus()

	if err != nil {
		return errors.Wrap(err, `拉取最新状态`)
	}

	for _, item := range *result {
		if singleErr := s.recordService.SetFinish(item.UID, item.Validate()); singleErr != nil {
			err = multierr.Append(err, singleErr) // 注意: 由于数据无法重复获取，因此拉取到的数据必须保存，这里不能直接返回
		}
	}

	return err
}
func (s Service) pullStatus() (result *pullStatusResult, err error) {
	values := url.Values{}
	values.Set(`apikey`, s.apiKey)
	values.Set(`page_size`, pageSize)

	result = &pullStatusResult{}

	if _, err := s.send(pullStatusURL, values, result, debug); err != nil {
		return nil, err
	}

	return result, nil
}

type pullStatusResult []pullSingleStatus

func (p pullStatusResult) Validate() error {
	return nil
}

type pullSingleStatus struct {
	ErrorDetail     string `json:"error_detail"`
	Sid             int    `json:"sid"`
	UID             string `json:"uid"`
	UserReceiveTime string `json:"user_receive_time"`
	ErrorMsg        string `json:"error_msg"`
	Mobile          string `json:"mobile"`
	ReportStatus    string `json:"report_status"`
}

func (p pullSingleStatus) success() bool {
	return p.ReportStatus == `SUCCESS`
}

func (p pullSingleStatus) Validate() error {
	if p.success() {
		return nil
	}

	if p.ErrorDetail != `` {
		return errors.New(p.ErrorDetail)
	}

	if p.ErrorMsg != `` {
		return errors.New(p.ErrorMsg)
	}

	return nil
}

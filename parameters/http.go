package parameters

import (
	"fmt"

	"github.com/fighterlyt/common/model/invoke"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (s *service) http() {
	s.router.POST(`/set`, s.setParameters)
	s.router.POST(`/get`, s.getParameters)
	s.router.POST(`/getHistory`, s.getHistory)
	s.router.POST(`/groupInfo`, s.httpGroupInfo)
}

func (s *service) httpGroupInfo(ctx *gin.Context) {
	invoke.ReturnSuccess(ctx, pageGroupInfo)
}

/*
setParameters 设置业务参数
参数:
*	ctx	*gin.Context	gin
返回值:
*/
func (s *service) setParameters(ctx *gin.Context) {
	argument := &setParametersArgument{}

	var (
		returned bool
		err      error
	)

	defer func() {
		if err != nil {
			s.logger.Error(err.Error())
		}
	}()

	if returned, err = invoke.ProcessArgument(ctx, argument); err != nil || returned {
		return
	}

	// 是否需要二次验证
	if s.needTwoFactor(argument.Parameters) {
		// 验证码为空,直接返回需要二次验证
		if argument.Code == "" {
			invoke.ReturnFail(ctx, invoke.NeedTwoFactor, errors.New("需要进行谷歌二次验证"), "请输入谷歌二次验证码")

			return
		}

		// 验证码不为空进行验证
		ok, err := s.auth.Validate(argument.Code)
		if !ok || err != nil {
			err = errors.New("验证码错误")
			invoke.ReturnFail(ctx, invoke.NeedTwoFactor, invoke.ErrFail, err.Error())

			return
		}
	}

	if err = s.Modify(argument.Parameters, argument.UserID); err != nil {
		err = errors.Wrap(err, `修改参数`)
		invoke.ReturnFail(ctx, invoke.Fail, invoke.ErrFail, err.Error())

		return
	}

	invoke.ReturnSuccess(ctx, nil)
}

func (s *service) needTwoFactor(keys map[string]string) (need bool) {
	for key, _ := range keys {
		for i := range s.needTwoFactorKeys {
			if key == s.needTwoFactorKeys[i] {
				return true
			}
		}
	}

	return false
}

type setParametersArgument struct {
	Parameters map[string]string `json:"parameters"`
	UserID     int64             `json:"userID"`
	Code       string            `json:"code"`
}

func (s setParametersArgument) Validate() error {
	if s.UserID <= 0 {
		return fmt.Errorf(`userID[%d]非法`, s.UserID)
	}

	if len(s.Parameters) == 0 {
		return errors.New(`参数不能为空`)
	}

	return nil
}

func (s *service) getParameters(ctx *gin.Context) {
	argument := &getParametersArgument{}

	var (
		returned   bool
		err        error
		parameters map[string]*Parameter
	)

	defer func() {
		if err != nil {
			s.logger.Error(err.Error())
		}
	}()

	if returned, err = invoke.ProcessArgument(ctx, argument); err != nil || returned {
		return
	}

	if parameters, err = s.GetParameters(argument.Keys...); err != nil {
		err = errors.Wrap(err, `获取参数`)
		invoke.ReturnFail(ctx, invoke.Fail, err, err.Error())

		return
	}
	for i := range parameters {
		if parameters[i] == nil {
			continue
		}

		if parameters[i].Hide {
			parameters[i].Value = ``
		}
	}

	invoke.ReturnSuccess(ctx, parameters)
}

type getParametersArgument struct {
	Keys []string `json:"keys"`
}

func (s *getParametersArgument) Validate() error {
	if len(s.Keys) == 0 {
		return errors.New(`参数不能为空`)
	}

	return nil
}

func (s *service) getHistory(ctx *gin.Context) {
	query := &historyQuery{}
	argument, err := invoke.NewListArgument(query)

	if err != nil {
		err = errors.Wrap(err, `构建列表参数错误`)
		invoke.ReturnFail(ctx, invoke.Fail, err, err.Error())

		return
	}

	var returned bool

	if returned, err = invoke.ProcessArgument(ctx, argument); err != nil || returned {
		return
	}

	var (
		result     []History
		allCount   int64
		listResult *invoke.ListResult
	)

	if allCount, result, err = s.GetHistory(query.Key, query.Start, query.End, argument.Start, argument.Limit); err != nil {
		err = errors.Wrap(err, `操作失败`)
		invoke.ReturnFail(ctx, invoke.Fail, err, err.Error())

		return
	}

	if listResult, err = invoke.NewListResult(allCount, result); err != nil {
		err = errors.Wrap(err, `构建列表返回值`)
		invoke.ReturnFail(ctx, invoke.Fail, err, err.Error())

		return
	}

	invoke.ReturnSuccess(ctx, listResult)
}

type historyQuery struct {
	Start int64  `json:"start"`
	End   int64  `json:"end"`
	Key   string `json:"key"`
}

func (d *historyQuery) Validate() error {
	if d.Start < 0 {
		return fmt.Errorf(`start[%d]必须大于等于0`, d.Start)
	}

	if d.End < 0 {
		return fmt.Errorf(`end[%d]必须大于等于0`, d.End)
	}

	if d.End <= d.Start && d.End != 0 {
		return fmt.Errorf(`end[%d]必须大于start[%d]`, d.End, d.Start)
	}

	// if d.Key == `` {
	// 	return errors.New(`key不能为空`)
	// }

	return nil
}

func (d *historyQuery) Scope(db *gorm.DB) *gorm.DB {
	return db
}

package options

import (
	"strings"

	"github.com/youthlin/t"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.com/nova_dubai/common/model/invoke"
)

func (s service) http() {
	s.IRoutes.POST(`/get`, s.httpGet)
}

type getArgument struct {
	Keys []string `json:"keys"`
}

func (g getArgument) Validate() error {
	if len(g.Keys) == 0 {
		return errors.New(`keys不能为空`)
	}

	for _, key := range g.Keys {
		if strings.TrimSpace(key) == `` {
			return errors.New(`key 不能为空`)
		}
	}

	return nil
}

func (s service) httpGet(ctx *gin.Context) {
	var (
		argument = &getArgument{}
		returned bool
		err      error
	)

	if returned, err = invoke.ProcessArgument(ctx, argument); returned {
		return
	}

	if err != nil {
		invoke.ReturnFail(ctx, invoke.Fail, invoke.ErrFail, err.Error())
		return
	}

	result := make(map[string][]item, len(argument.Keys))

	for _, key := range argument.Keys {
		result[key] = Get(key)
	}

	var ts *t.Translations

	if tr, ok := ctx.Get("$Translations"); ok {
		s.Logger.Warn(`多语言`, zap.Any(`tr`, tr), zap.Reflect(`tr`, tr))

		if ts, ok = tr.(*t.Translations); ok {
			s.Logger.Warn(`多语言`, zap.Any(`locale`, ts.Locale()), zap.Any(`domains`, ts.GetOrNoop(`default`)))
			for key, item := range result {
				for i := range item {
					result[key][i].Text = ts.T(result[key][i].Text)
				}
			}
			s.Logger.Warn(`多语言结果`, zap.Any(`结果`, result))
		}
	}

	invoke.ReturnSuccess(ctx, result)
}

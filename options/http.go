package options

import (
	"strings"

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

	invoke.ReturnSuccess(ctx, result)
}

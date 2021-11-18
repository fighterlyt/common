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
	Key string `json:"key"`
}

func (g getArgument) Validate() error {
	if strings.TrimSpace(g.Key) == `` {
		return errors.New(`key 不能为空`)
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

	invoke.ReturnSuccess(ctx, Get(argument.Key))
}

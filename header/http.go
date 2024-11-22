package header

import (
	"strings"

	"github.com/fighterlyt/common/model/invoke"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (s service) http() {
	s.IRoutes.POST("/get", s.httpGet)
}

type Resp struct {
	ResourceType string      `json:"resource_type"`
	Resource     interface{} `json:"resource"`
}

func NewDefaultHeader(header DefaultHeader) Resp {
	return Resp{
		ResourceType: "default",
		Resource:     header,
	}
}

func NewSortableHeader(header SortableHeader) Resp {
	return Resp{
		ResourceType: "sortable",
		Resource:     header,
	}
}

func NewProjectHeader(header ProjectHeader) Resp {
	return Resp{
		ResourceType: "project",
		Resource:     header,
	}
}

func NewMultipleHeader(header MultipleHeader) Resp {
	return Resp{
		ResourceType: "multiple",
		Resource:     header,
	}
}

type DefaultHeader = Header
type SortableHeader = Header
type ProjectHeader = Header

type MultipleHeader struct {
	Column []Header `json:"column"`
}

type Header struct {
	Prop        string `json:"prop"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type Multiple struct {
	Column []Header `json:"column"`
}

type headerGetArgument struct {
	Keys []string `json:"keys"`
}

func (h *headerGetArgument) Validate() error {
	if len(h.Keys) != 1 {
		return errors.New(`keys长度只能为1`)
	}

	for _, key := range h.Keys {
		if strings.TrimSpace(key) == `` {
			return errors.New(`key 不能为空`)
		}
	}

	return nil
}

func (s *service) httpGet(ctx *gin.Context) {
	v := new(headerGetArgument)

	returned, err := invoke.ProcessArgument(ctx, v)
	if returned {
		return
	}

	if err != nil {
		invoke.ReturnFail(ctx, invoke.Fail, invoke.ErrFail, err.Error())
		return
	}

	invoke.ReturnSuccess(ctx, Get(v.Keys[0]))
}

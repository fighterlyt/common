package helpers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

const (
	spanKey = `span`
)

func PutSpanInGin(span opentracing.Span, ctx *gin.Context) {
	ctx.Set(spanKey, span)
}

func GetSpanFromGin(ctx *gin.Context) *Span {
	value := ctx.Value(spanKey)

	if value == nil {
		return nil
	}

	span, ok := value.(opentracing.Span)

	if ok {
		return &Span{span: span}
	}

	return nil
}

func PutSpanInCtx(span *Span, parent context.Context) context.Context {
	return context.WithValue(parent, spanKey, span)
}

func GetSpanFromCtx(ctx context.Context) *Span {
	value := ctx.Value(spanKey)

	if value == nil {
		return nil
	}

	span, ok := value.(*Span)

	if ok {
		return span
	}

	return nil
}

type Span struct {
	span opentracing.Span
}

func StartChild(span *Span, name string) *Span {
	if span == nil {
		return nil
	}

	return &Span{span: span.span.Tracer().StartSpan(name, opentracing.ChildOf(span.span.Context()))}
}

func (s *Span) FinishSpan(err error) {
	if s == nil {
		return
	}

	errMsg := ``

	if err != nil {
		errMsg = err.Error()
	}

	s.span.SetTag(`err`, errMsg)
	s.span.Finish()
}

func (s *Span) SetTag(key string, value interface{}) {
	if s == nil {
		return
	}

	s.span.SetTag(key, value)
}

type GormTracing struct {
}

func (g GormTracing) BeforeFind(tx *gorm.DB) {
	ctx := tx.Statement.Context

	if ctx == nil {
		return
	}

	span := StartChild(GetSpanFromCtx(ctx), `mysql-query`)

	tx.Statement.Context = PutSpanInCtx(span, ctx)
}

func (g GormTracing) AfterFind(tx *gorm.DB) {
	ctx := tx.Statement.Context

	if ctx == nil {
		return
	}

	span := GetSpanFromCtx(ctx)
	if span == nil {
		return
	}

	span.SetTag(`stmt`, tx.Dialector.Explain(tx.Statement.SQL.String(), tx.Statement.Vars...))

	span.FinishSpan(tx.Error)
}

package helpers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

const (
	spanKey SpanKey = `span`
)

type SpanKey string

/*PutSpanInGin span放入gin
参数:
*	span	opentracing.Span	链路追踪块
*	ctx 	*gin.Context    	gin上下文
返回值:
*/
func PutSpanInGin(span opentracing.Span, ctx *gin.Context) {
	ctx.Set(string(spanKey), span)
}

/*GetSpanFromGin 从gin上下文中获取链路追踪块
参数:
*	ctx  	*gin.Context	gin上下文
返回值:
*	*Span	*Span       	链路追踪块
*/
func GetSpanFromGin(ctx *gin.Context) *Span {
	value := ctx.Value(string(spanKey))

	if value == nil {
		return nil
	}

	span, ok := value.(opentracing.Span)

	if ok {
		return &Span{span: span}
	}

	return nil
}

/*PutSpanInCtx span放入上下文
参数:
*	span           	*Span          	链路追踪块
*	parent         	context.Context	上下文
返回值:
*	context.Context	context.Context	放入链路追踪块之后的上下文
*/
func PutSpanInCtx(span *Span, parent context.Context) context.Context { // nolint:revive
	return context.WithValue(parent, spanKey, span)
}

/*GetSpanFromCtx 从上下文中获取链路追踪块
参数:
*	ctx  	context.Context	上下文
返回值:
*	*Span	*Span          	链路追踪块
*/
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

/*StartChild 创建启动并返回一个包含上一个的span的新span
参数:
*	span 	*Span 	原span
*	name 	string	新span名字
返回值:
*	*Span	*Span 	新span
*/
func StartChild(span *Span, name string) *Span {
	if span == nil {
		return nil
	}

	return &Span{span: span.span.Tracer().StartSpan(name, opentracing.ChildOf(span.span.Context()))}
}

/*FinishSpan 	 设置结束时间戳并最终确定 Span 状态
参数:
*	err	error	如果有错误会把错误携带上
返回值:
*/
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

/*SetTag 设置标签
参数:
*	key  	string 标签键
*	value	interface{}	标签值
返回值:
*/
func (s *Span) SetTag(key string, value interface{}) {
	if s == nil {
		return
	}

	s.span.SetTag(key, value)
}

type GormTracing struct {
}

/*BeforeFind gorm查询前钩子设置span
参数:
*	tx	*gorm.DB	gorm
返回值:
*/
func (g GormTracing) BeforeFind(tx *gorm.DB) {
	ctx := tx.Statement.Context

	if ctx == nil {
		return
	}

	span := StartChild(GetSpanFromCtx(ctx), `mysql-query`)

	tx.Statement.Context = PutSpanInCtx(span, ctx)
}

/*AfterFind gorm查询后钩子把生成的sql语句放入tag中
参数:
*	tx	*gorm.DB	gorm
返回值:
*/
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

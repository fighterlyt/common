package helpers

import (
	"context"
	"io"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap/buffer"
	"gorm.io/gorm"
)

const (
	spanKey    SpanKey = `span`
	spanGinKey string  = `span-gin`
)

func loadTracer(serviceName string) (tracer *Tracer, err error) {
	cfg := jaegercfg.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
	}

	tracer = &Tracer{}

	if tracer.Tracer, tracer.Closer, err = cfg.NewTracer(); err != nil {
		return nil, errors.Wrap(err, `初始化tracer`)
	}

	return tracer, nil
}

type Tracer struct {
	Tracer opentracing.Tracer
	Closer io.Closer
}

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

func NewSpan(operationName string, tags map[string]interface{}) *Span {
	span := opentracing.StartSpan(operationName, opentracing.StartTime(time.Now().In(GetDefaultLocation())), opentracing.Tags(tags))
	return &Span{span: span}
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

// GormTracing gorm 追踪
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

func getBody(ctx *gin.Context) string {
	data := &buffer.Buffer{}

	if ctx.Request.Body != nil {
		_, _ = io.Copy(data, ctx.Request.Body)
	}

	body := data.String()
	displayBody := body

	if len(body) > 4*1024 {
		displayBody = `大于4K,请查看日志`
	}

	ctx.Request.Body = io.NopCloser(strings.NewReader(body))

	return displayBody
}

var once sync.Once

func Trace(tracer opentracing.Tracer, ignorePrefix ...string) func(ctx *gin.Context) {
	once.Do(func() {
		opentracing.SetGlobalTracer(tracer)
	})

	return func(ctx *gin.Context) {
		if tracer != nil {
			visit := ctx.Request.URL.String()

			for _, prefix := range ignorePrefix {
				if strings.HasPrefix(visit, prefix) {
					visit = prefix
					break
				}
			}

			start := time.Now()

			span := tracer.StartSpan(visit, opentracing.StartTime(start), opentracing.Tags{
				`token`:      ctx.GetHeader(`token`),
				`method`:     ctx.Request.Method,
				`path`:       ctx.Request.URL.Path,
				`query`:      ctx.Request.URL.RawQuery,
				`ip`:         ctx.ClientIP(),
				`user-agent`: ctx.Request.UserAgent(),
				`req-size`:   ctx.Request.ContentLength,
			})

			span.SetTag(`body`, getBody(ctx))

			PutSpanInGin(span, ctx)
			ctx.Next()

			duration := time.Since(start)

			setPostValues(span, ctx)

			span.SetTag(`latency`, duration.String())
			span.SetTag(`status`, ctx.Writer.Status())
			span.SetTag(`size`, ctx.Writer.Size())
			span.Finish()
		}
	}
}

func setPostValues(span opentracing.Span, ctx *gin.Context) {
	value := ctx.Value(spanGinKey)

	if m, ok := value.(map[string]interface{}); ok {
		for k, v := range m {
			span.SetTag(spanGinKey+`_`+k, v)
		}
	} else if value != nil {
		span.SetTag(`type`, reflect.TypeOf(value).String())
	}
}

/*GinTraceSet 将数据写入到span 中，注意value类型必须是字符串、数值和布尔
参数:
*	ctx  	*gin.Context	gin环境变量
*	key  	string          key
*	value	interface{} 	值,必须是字符串、数值、布尔类型
返回值:
*/
func GinTraceSet(ctx *gin.Context, key string, value interface{}) {
	spanValue := ctx.Value(spanGinKey)
	if spanValue == nil {
		spanValue = make(map[string]interface{})
	}

	if m, ok := spanValue.(map[string]interface{}); ok {
		m[key] = value
		ctx.Set(spanGinKey, m)
	}
}

/*WrapGinMiddle 封装gin的中间件，提供了时间信息
参数:
*	name                  	string                	中间件名称
*	fun                   	func(ctx *gin.Context)	中间件方法
返回值:
*	func(ctx *gin.Context)	func(ctx *gin.Context)	封装后的中间件
*/
func WrapGinMiddle(name string, fun func(ctx *gin.Context)) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		span := StartChild(GetSpanFromGin(ctx), `gin-`+name)

		fun(ctx)
		ctx.Next()

		span.FinishSpan(nil)
	}
}

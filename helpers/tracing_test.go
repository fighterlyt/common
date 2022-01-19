package helpers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

var (
	engine *gin.Engine
	tracer *Tracer
	err    error
)

func TestTracingBase(t *testing.T) {
	tracer, err = loadTracer(`test`)
	require.NoError(t, err, `loadTracer`)

	engine = gin.Default()
	engine.Use(Trace(tracer.Tracer))
}

func TestTracingAfter(t *testing.T) {
	engine.GET(`/test`, func(context *gin.Context) {
		context.String(http.StatusOK, `ok`)
	})

	go func() {
		engine.Run()
	}()

	var (
		resp *http.Response
	)

	resp, err = http.Get(`http://localhost:8080/test`)
	require.NoError(t, err)

	_ = resp.Body.Close()

	time.Sleep(time.Second * 5)
}
func TestStartChild(t *testing.T) {
	TestTracingBase(t)

	engine.Use(func(context *gin.Context) {
		span := StartChild(GetSpanFromGin(context), `testChild`)
		time.Sleep(time.Millisecond * 10)
		span.FinishSpan(nil)
		t.Log(`use`)
		context.Next()
	})

	TestTracingAfter(t)
}

func TestGinTraceSet(t *testing.T) {
	TestTracingBase(t)
	engine.Use(func(context *gin.Context) {
		GinTraceSet(context, `test`, `1`)
		context.Next()
	})

	TestTracingAfter(t)
}

func TestGinWrap(t *testing.T) {
	TestTracingBase(t)
	engine.Use(WrapGinMiddle(`测试`, func(context *gin.Context) {
		GinTraceSet(context, `test`, `1`)
		context.Next()
	}))

	TestTracingAfter(t)
}

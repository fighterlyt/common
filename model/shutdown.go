package model

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/nova_dubai/common/model/invoke"
	"go.uber.org/atomic"
)

type Shutdown interface {
	Close()
	Add(count int64)
	IsClosed() bool
	IsFinished() bool
}

// GinShutdown 是控制gin HTTP的关闭
type GinShutdown struct {
	*shutdown
}

/*NewGinShutdown 新建
参数:
返回值:
*	*GinShutdown	*GinShutdown	数据
*/
func NewGinShutdown() *GinShutdown {
	return &GinShutdown{
		shutdown: NewShutdown(),
	}
}

/*Process 处理
参数:
*	ctx	*gin.Context	上下文
返回值:
*/
func (g *GinShutdown) Process(ctx *gin.Context) {
	if g.closed {
		ctx.AbortWithStatusJSON(http.StatusNotFound, invoke.NewResult(invoke.Fail, `服务器已经关闭`, nil, `服务器已经关闭`))

		return
	}

	g.Add(1)

	ctx.Next()

	g.Add(-1)
}

// shutdown 关闭控制器
type shutdown struct {
	counter *atomic.Int64
	closed  bool
}

/*NewShutdown 新建控制器
参数:
返回值:
*	*shutdown	*shutdown	返回值1
*/
func NewShutdown() *shutdown {
	return &shutdown{
		counter: atomic.NewInt64(0),
	}
}

/*Close 关闭
参数:
返回值:
*/
func (s *shutdown) Close() {
	s.closed = true
}

/*Add 添加计数
参数:
*	count	int64	参数1
返回值:
*/
func (s *shutdown) Add(count int64) {
	s.counter.Add(count)
}

/*IsClosed 是否已经关闭
参数:
返回值:
*	bool	bool	返回值1
*/
func (s *shutdown) IsClosed() bool {
	return s.closed
}

func (s *shutdown) IsFinished() bool {
	return s.counter.Load() == 0 && s.closed
}

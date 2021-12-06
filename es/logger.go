package es

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fighterlyt/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type elasticLogger struct {
	log.Logger
	level zapcore.Level
}

func (e elasticLogger) LogRoundTrip(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) error {
	output := bytes.Buffer{}

	fmt.Fprintf(&output, "%s %s %s [status:%d request:%s]\n",
		start.Format(time.RFC3339),
		req.Method,
		req.URL.String(),
		resStatusCode(res),
		dur.Truncate(time.Millisecond),
	)

	if e.RequestBodyEnabled() && req != nil && req.Body != nil && req.Body != http.NoBody {
		var buf bytes.Buffer

		if req.GetBody != nil {
			b, _ := req.GetBody()
			_, _ = buf.ReadFrom(b)
		} else {
			_, _ = buf.ReadFrom(req.Body)
		}

		logBodyAsText(&output, &buf, ">")
	}

	if e.ResponseBodyEnabled() && res != nil && res.Body != nil && res.Body != http.NoBody {
		defer func() {
			_ = res.Body.Close()
		}()

		var buf bytes.Buffer

		_, _ = buf.ReadFrom(res.Body)
		logBodyAsText(&output, &buf, "<")
	}

	if err != nil {
		_, _ = fmt.Fprintf(&output, "! ERROR: %v\n", err)
	}

	e.Logger.Info(`es日志`, zap.String(`内容`, output.String()))

	return nil
}

func resStatusCode(res *http.Response) int {
	if res == nil {
		return -1
	}

	return res.StatusCode
}

func (e elasticLogger) RequestBodyEnabled() bool {
	return e.level == zapcore.DebugLevel
}

func (e elasticLogger) ResponseBodyEnabled() bool {
	return e.level == zapcore.DebugLevel
}

func newElasticLogger(logger log.Logger, level zapcore.Level) *elasticLogger {
	return &elasticLogger{Logger: logger, level: level}
}

func (e elasticLogger) Printf(format string, v ...interface{}) {
	switch e.level {
	case zapcore.DebugLevel:
		e.Debug(fmt.Sprintf(format, v...))
	case zapcore.InfoLevel:
		e.Info(fmt.Sprintf(format, v...))
	case zapcore.WarnLevel:
		e.Warn(fmt.Sprintf(format, v...))
	default:
	}
}

func logBodyAsText(dst io.Writer, body io.Reader, prefix string) {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		s := scanner.Text()
		if s != "" {
			fmt.Fprintf(dst, "%s %s\n", prefix, s)
		}
	}
}

package es

import (
	"bytes"
	"context"
	"encoding/json"

	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/fighterlyt/log"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
)

// Client 客户端
type Client struct {
	*es.Client
	logger log.Logger
}

// ClientArgument 客户端参数
type ClientArgument struct {
	logger  log.Logger // 日志器
	httpURL []string   // http 地址
	user    string     // 账号
	pass    string     // 密码
}

/*NewClientArgument 新建客户端参数
参数:
*	logger         	log.Logger     	日志器
*	httpURL        	[]string        访问地址
*	user           	string         	参数3
*	pass           	string         	参数4
返回值:
*	*ClientArgument	*ClientArgument	返回值1
*/
func NewClientArgument(logger log.Logger, httpURL []string, user, pass string) *ClientArgument {
	return &ClientArgument{
		logger:  logger,
		httpURL: httpURL,
		user:    user,
		pass:    pass,
	}
}

func NewClient(argument *ClientArgument) (client *Client, err error) {
	var (
		originClient *es.Client
	)

	cfg := es.Config{
		Addresses:             argument.httpURL,
		Username:              argument.user,
		Password:              argument.pass,
		CloudID:               "",
		APIKey:                "",
		Header:                nil,
		CACert:                nil,
		RetryOnStatus:         nil,
		DisableRetry:          false,
		EnableRetryOnTimeout:  false,
		MaxRetries:            0,
		DiscoverNodesOnStart:  false,
		DiscoverNodesInterval: 0,
		EnableMetrics:         false,
		EnableDebugLogger:     false,
		RetryBackoff:          nil,
		Transport:             nil,
		Logger:                newElasticLogger(argument.logger, zapcore.DebugLevel),
		Selector:              nil,
		ConnectionPoolFunc:    nil,
	}

	if originClient, err = es.NewClient(cfg); err != nil {
		return nil, errors.Wrap(err, `连接错误`)
	}

	return &Client{
		Client: originClient,
		logger: argument.logger,
	}, nil
}

func (c *Client) CreateIndex(name string) error {
	// indexService := c.Client.CreateIndex(name)
	//
	// result, err := indexService.Do(bg)
	//
	// if err != nil {
	// 	return errors.Wrap(err, `创建索引错误`)
	// }
	//
	// if !result.Acknowledged {
	// 	return errors.New(`未完成`)
	// }

	return nil
}

func (c *Client) Add(document Document) error {
	var (
		jsonDocument []byte
		err          error
		resp         *esapi.Response
	)

	if jsonDocument, err = json.Marshal(document); err != nil {
		return errors.Wrap(err, `序列化错误`)
	}

	req := esapi.IndexRequest{
		Index:               document.Index(),
		DocumentID:          document.GetID(),
		Body:                bytes.NewReader(jsonDocument),
		IfPrimaryTerm:       nil,
		IfSeqNo:             nil,
		OpType:              "",
		Pipeline:            "",
		Refresh:             "",
		RequireAlias:        nil,
		Routing:             "",
		Timeout:             0,
		Version:             nil,
		VersionType:         "",
		WaitForActiveShards: "",
		Pretty:              false,
		Human:               false,
		ErrorTrace:          false,
		FilterPath:          nil,
		Header:              nil,
	}

	resp, err = req.Do(bg, c.Client)

	if err != nil {
		return errors.Wrap(err, `创建索引错误`)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.IsError() {
		return errors.New(resp.String())
	}

	return nil
}

var (
	bg = context.Background()
)

type Document interface {
	Index() string
	GetID() string
}

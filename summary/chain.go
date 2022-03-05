package summary

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"go.uber.org/multierr"
)

/*Chain 将多个Client 连接
参数:
*	clients	...Client	需要连接的客户端
返回值:
*	client 	Client   	连接后的客户端
*	err    	error    	错误
*/
func Chain(clients ...Client) (client Client, err error) {
	exists := make(map[string]struct{}, len(clients))
	result := make([]Client, 0, len(clients))

	// 去重
	for i, elem := range clients {
		if elem == nil {
			return nil, errors.Wrapf(err, `参数不能为空,第[%d]个参数为空`, i+1)
		}

		if _, exist := exists[elem.Key()]; !exist {
			exists[elem.Key()] = struct{}{}

			result = append(result, clients[i])
		}
	}

	return chainClients(result), nil
}

type chainClients []Client

func (c chainClients) Model() Summary {
	return nil
}

func (c chainClients) Summarize(ownerID int64, amount decimal.Decimal) error {
	var (
		err, singleErr error
	)

	for _, client := range c {
		if singleErr = client.Summarize(ownerID, amount); singleErr != nil {
			err = multierr.Append(err, errors.Wrap(singleErr, client.Key()))
		}
	}

	return err
}

func (c chainClients) SummarizeDay(date int64, ownerID string, amount decimal.Decimal) error {
	return nil
}
func (c chainClients) Key() string {
	build := &strings.Builder{}

	build.WriteString(`chain`)

	for _, client := range c {
		build.WriteString(`_` + client.Key())
	}

	return build.String()
}

func (c chainClients) GetSummary(ownerIDs []string, from, to int64) (records []Summary, err error) {
	return nil, errors.New(`chain不支持查询`)
}

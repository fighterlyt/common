package summaryextend

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

func (c chainClients) RevertSummarizeDay(date int64, ownerID string, amount decimal.Decimal, extendValue ...decimal.Decimal) error {
	var (
		err, singleErr error
	)

	for _, client := range c {
		if singleErr = client.RevertSummarizeDay(date, ownerID, amount, extendValue...); singleErr != nil {
			err = multierr.Append(err, errors.Wrap(singleErr, client.Key()))
		}
	}

	return err
}

func (c chainClients) Model() Summary {
	return nil
}

func (c chainClients) Summarize(ownerID string, amount decimal.Decimal, extend ...decimal.Decimal) error {
	var (
		err, singleErr error
	)

	for _, client := range c {
		if singleErr = client.Summarize(ownerID, amount, extend...); singleErr != nil {
			err = multierr.Append(err, errors.Wrap(singleErr, client.Key()))
		}
	}

	return err
}

func (c chainClients) SummarizeDay(_ int64, _ string, _ decimal.Decimal, _ ...decimal.Decimal) error {
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

func (c chainClients) GetSummary(_ []string, _, _ int64) (records []Summary, err error) {
	return nil, errors.New(`chain不支持查询`)
}
func (c chainClients) GetSummaryByLike(_ string, _, _ int64) (records []Summary, err error) {
	return nil, errors.New(`chain不支持查询`)
}

func (c chainClients) GetSummarySummary(ownerIDs []string, from, to int64) (record Summary, err error) {
	return nil, errors.New(`chain不支持查询`)
}

func (c chainClients) GetSummaryExclude(_ []string, _, _ int64, _ ...string) (records []Summary, err error) {
	return nil, errors.New(`chain 不支持查询`)
}

package helpers

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

/*RedisDelInBatch 批量执行REDIS DEL 命令，主要是防止一次性传入的KEY 太多，REDIS 故障
参数:
*	ctx      	context.Context	上下文
*	client   	*redis.Client   redis 客户端
*	keys     	[]string       	需要删除的key
*	batchSize	int            	批量数量，必须大于0
返回值:
*	error    	error          	错误
*/
func RedisDelInBatch(ctx context.Context, client *redis.Client, keys []string, batchSize int) error {
	if batchSize < 1 {
		return fmt.Errorf(`batchSize 必须大于 0，实际为[%d]`, batchSize)
	}

	start := 0
	end := start + batchSize

	for {
		if end > len(keys) {
			end = len(keys)
		}

		if err := client.Del(ctx, keys[start:end]...).Err(); err != nil {
			return errors.Wrapf(err, `REDIS DEL %s`, strings.Join(keys[start:end], ` `))
		}

		start = end
		end += batchSize

		if start == len(keys) {
			break
		}
	}

	return nil
}

package id

import (
	"github.com/bwmarrin/snowflake"
	"github.com/pkg/errors"
)

type generator struct {
	node *snowflake.Node
}

/*Generate 生成ID
参数:
返回值:
*	snowflake.ID	snowflake.ID	生成ID
*/
func (g generator) Generate() snowflake.ID {
	return g.node.Generate()
}

/*NewGenerator 生成构造器
参数:
*	id    	int64    	节点ID
返回值:
*	target	Generator	构造器
*	err   	error    	错误
*/
func NewGenerator(id int64) (target Generator, err error) {
	var (
		node *snowflake.Node
	)

	if node, err = snowflake.NewNode(id); err != nil {
		return nil, errors.Wrap(err, `生成snowflake失败`)
	}

	return &generator{node: node}, nil
}

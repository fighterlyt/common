package system

import (
	"fmt"
	"syscall"

	"github.com/pkg/errors"
)

/*SetNoFileLimit 设置打开文件
参数:
*	count	uint64  数量
返回值:
*	error	error
*/
func SetNoFileLimit(count uint64) error {
	rLimit := syscall.Rlimit{
		Max: count,
		Cur: count,
	}

	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return errors.Wrap(err, "设置ulimit")
	}

	if err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return errors.Wrap(err, "设置ulimit")
	}

	if err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return errors.Wrap(err, "获取ulimit")
	}

	if rLimit.Cur != count {
		return fmt.Errorf("cur值不对,应该是[%d],实际是[%d],完整值%d/%d", count, rLimit.Cur, rLimit.Cur, rLimit.Max)
	}

	if rLimit.Max != count {
		return fmt.Errorf("max值不对,应该是[%d],实际是[%d]", count, rLimit.Max)
	}

	return nil
}

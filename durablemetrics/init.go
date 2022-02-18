package durablemetrics

var (
	argument InitArgument
)

/*Init 初始化
参数:
*	initArgument	InitArgument	初始化参数
返回值:
*/
func Init(initArgument InitArgument) {
	argument = initArgument
}

// InitArgument 初始化参数
type InitArgument struct {
	Namespace string // 系统名
}

package mongo

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type QueryFilter struct {
	query     bson.M             // 查询条件
	sort      []string           // 排序规则
	start     int64              // 开始位置
	limit     int64              // 数量
	include   []string           // 包含字段
	exclude   []string           // 排除字段
	data      interface{}        // 单个对象，非指针
	all       bool               // 是否返回全部数量
	collation *options.Collation // collation规则
}

func NewQueryFilter(query bson.M, sort []string, start, limit int64, include, exclude []string, data interface{}, all bool, collation *options.Collation) *QueryFilter { //nolint:lll
	return &QueryFilter{
		query:     query,
		sort:      sort,
		start:     start,
		limit:     limit,
		include:   include,
		exclude:   exclude,
		data:      data,
		all:       all,
		collation: collation,
	}
}

func (q *QueryFilter) Query() bson.M {
	return q.query
}

func (q *QueryFilter) SetQuery(query bson.M) {
	q.query = query
}

func (q *QueryFilter) Sort() []string {
	return q.sort
}

func (q *QueryFilter) SetSort(sort []string) {
	q.sort = sort
}

func (q *QueryFilter) Start() int64 {
	return q.start
}

func (q *QueryFilter) SetStart(start int64) {
	q.start = start
}

func (q *QueryFilter) Limit() int64 {
	return q.limit
}

func (q *QueryFilter) SetLimit(limit int64) {
	q.limit = limit
}

func (q *QueryFilter) Include() []string {
	return q.include
}

func (q *QueryFilter) SetInclude(include []string) {
	q.include = include
}

func (q *QueryFilter) Exclude() []string {
	return q.exclude
}

func (q *QueryFilter) SetExclude(exclude []string) {
	q.exclude = exclude
}

func (q *QueryFilter) Data() interface{} {
	return q.data
}

func (q *QueryFilter) SetData(data interface{}) {
	q.data = data
}

func (q *QueryFilter) All() bool {
	return q.all
}

func (q *QueryFilter) SetAll(all bool) {
	q.all = all
}

func (q *QueryFilter) Collation() *options.Collation {
	return q.collation
}

func (q *QueryFilter) SetCollation(collations *options.Collation) {
	q.collation = collations
}

func (q QueryFilter) MakeSelect() (selection bson.M, err error) {
	return MakeSelect(q.include, q.exclude)
}

func (q QueryFilter) FindOptions(selection bson.M) *options.FindOptions {
	return &options.FindOptions{
		Sort:       convertSort(q.sort),
		Limit:      &q.limit,
		Skip:       &q.start,
		Projection: selection,
		Collation:  q.collation,
	}
}

func (q QueryFilter) CountOptions() *options.CountOptions {
	return &options.CountOptions{
		Collation: q.collation,
	}
}

/*Query 查询
参数:
*	ctx       	context.Context		context
*	collection	*mongo.Collection	集合
*	query     	bson.M				查询条件
*	sort      	[]string			排序规则
*	start     	int64				开始顺序
*	limit     	int64				数量
*	include   	[]string			包含字段
*	exclude   	[]string			排除字段
*	data      	interface{}			单个对象,非指针
*	all       	bool				是否返回全部数量
*   collations []options.Collation  collation规则，最多只有1个
返回值:
*	result	[]interface{}
*	count 	int64
*	err   	error
*/
func Query(ctx context.Context, collection *mongo.Collection, filter *QueryFilter) (result []interface{}, count int64, err error) {
	var (
		selection bson.M
		cursor    *mongo.Cursor
	)

	if selection, err = filter.MakeSelect(); err != nil {
		return nil, 0, err
	}

	if len(selection) == 0 {
		selection = nil
	}

	option := filter.FindOptions(selection)

	if cursor, err = collection.Find(ctx, filter.Query(), option); err != nil {
		return nil, 0, errors.Wrap(err, "find")
	}

	t := reflect.TypeOf(filter.Data())

	for cursor.Next(ctx) {
		count++

		element := reflect.New(t)

		if err = cursor.Decode(element.Interface()); err != nil {
			return nil, 0, errors.Wrap(err, "bson解码")
		}

		result = append(result, element.Elem().Interface())
	}

	if filter.All() {
		countOption := filter.CountOptions()

		if count, err = collection.CountDocuments(ctx, filter.query, countOption); err != nil {
			return nil, 0, errors.Wrap(err, "计数")
		}
	}

	return result, count, nil
}

/*MakeSelect 根据参数生成对应的bson.M,用于(*mgo.Query).Select()参数
参数:
*	include	[]string
*	exclude	[]string
返回值:
*	bson.M	bson.M
*	error 	error
*/
func MakeSelect(include, exclude []string) (bson.M, error) {
	if err := validateSelect(include, exclude); err != nil {
		return nil, err
	}

	return combineSelect(include, exclude), nil
}

func combineSelect(include, exclude []string) bson.M {
	fields := bson.M{}
	for _, field := range include {
		fields[field] = 1
	}

	for _, field := range exclude {
		fields[field] = 0
	}

	return fields
}

func validateSelect(include, exclude []string) error {
	if len(include) != 0 && len(exclude) != 0 {
		return errors.New("两个参数必须至少有一个为空")
	}

	return nil
}

/*convertSort 转换排序条件，排序中-xxx,表示倒序
参数:
*	sorts 	[]string	字符串表示的顺序
返回值:
*	bson.D	bson.D  	适合mongo的排序规则
*/
func convertSort(sorts []string) bson.D {
	result := make([]bson.E, 0, len(sorts))

	var (
		value int
	)

	for _, sort := range sorts {
		if sort[0:1] == "-" {
			sort = sort[1:]
			value = -1
		} else {
			value = 1
		}

		result = append(result, bson.E{
			Key:   sort,
			Value: value,
		})
	}

	return result
}

// QuerySpec 查询条件，外部数据->数据使用规则
type QuerySpec map[string]DBSpec

// DBSpec 数据使用规则
type DBSpec struct {
	Field   string   // 针对的字段
	Fields  []string // 针对的多个字段
	Dynamic bool     // 是否动态,结果是条件bson.M，而不是值
	Op      string   // 操作符
	Convert Convert  // 转化规则(传入的查询值->数据库值),动态时，返回的是一个bson.M
}

// Convert 转换规则,将字符串转为对应的类型
type Convert func(data string) (interface{}, error)

func buildStaticQuery(input map[string]interface{}, static map[string][]*querySpec) (result map[string]interface{}, used map[string]struct{}, err error) { //nolint:lll
	var (
		value interface{}
		exist bool
	)

	result = make(map[string]interface{}, len(input)) // 结果
	used = make(map[string]struct{}, len(input))      // 记录使用过的数据

	for field, spec := range static { // 遍历规则，field 是字段，spec 是该字段全部规则
		for _, s := range spec { // 单条规则
			if value, exist = input[s.key]; !exist { // 如果规则能够适用，也就是有对应的输入值
				continue
			}

			if value, err = s.GetValue(value.(string)); err != nil {
				return nil, nil, fmt.Errorf("转换数据发生问题:数据:%#v 类型:%T 字段:%s", input[s.key], input[s.key], s.key)
			}

			if _, exist = result[field]; !exist {
				result[field] = make(map[string]interface{}, defaultCapacity)
			}

			switch len(spec) { // 规则的数量
			case 0: // 没有规则
				delete(result, field)
			case 1: // 单个规则
				switch s.op {
				case REGEX:
					result[field] = bson.M{s.op: value, "$options": "i"} // 正则表达式，忽略大小写
				case IN, NIN:
					if value != nil {
						reflectValue := reflect.ValueOf(value)

						if reflectValue.Kind() != reflect.Slice {
							return nil, nil, fmt.Errorf("[%s]操作符对应的值必须是slice,对应字段[%s]", s.op, s.key)
						}

						if !reflectValue.IsNil() { // 非空的值，参加，这里存在interface{}不为空，但是interface{}的值是空
							result[field] = bson.M{s.op: value}
						}
					}
				default:
					result[field] = bson.M{s.op: value}
				}
			default:
				result[field].(map[string]interface{})[s.op] = value

				if s.op == REGEX {
					result[field].(map[string]interface{})["$options"] = "i"
				}
			}

			used[s.key] = struct{}{} // 记录已经使用过的数据，用过的数据不能直接删除，可能会多次使用
		}
	}

	return result, used, nil
}

func buildDynamicQuery(input map[string]interface{}, dynamic map[string]Convert) (result map[string]interface{}, used map[string]struct{}, err error) { //nolint:lll
	var (
		value interface{}
		exist bool
		field primitive.M
		ok    bool
	)

	result = make(map[string]interface{}, len(input)) // 结果
	used = make(map[string]struct{}, len(input))      // 记录使用过的数据

	for inputKey, convert := range dynamic { // 遍历动态规则，key是输入的key
		if value, exist = input[inputKey]; !exist { // 外部传入值
			continue
		}

		if value, err = convert(value.(string)); err != nil {
			return nil, nil, fmt.Errorf("转换数据发生问题:\n\t数据:%#v\n\t类型:%T\n\t字段:%s", input[inputKey], reflect.TypeOf(input[inputKey]).String(), inputKey)
		}

		if field, ok = value.(primitive.M); !ok {
			return nil, nil, fmt.Errorf("动态字段[%s]Convert第一个返回值必须是bson.M,实际上是[%s]", inputKey, reflect.TypeOf(value).String())
		}

		for key, condition := range field { // 遍历生成的动态条件，key可能是mongodb 操作符，也有可能是其他的
			switch key {
			case OR, AND:
				// todo: 如果有多个$or或者$and 需要汇总
				result[key] = condition
			default:
				if result[key] == nil { // 补充空栏位
					result[key] = make(map[string]interface{}, defaultCapacity)
				}

				if fullCondition, ok := condition.(primitive.M); ok {
					for operator, value := range fullCondition { // 这里不能覆盖，只能附加
						result[key].(map[string]interface{})[operator] = value
					}

					continue
				}

				switch t := result[key].(type) {
				case primitive.M:
					t[EQ] = condition
				case map[string]interface{}:
					result[key].(map[string]interface{})[EQ] = condition
				}
			}
		}

		delete(input, inputKey)
	}

	return result, used, nil
}

/*BuildQuery 构建查询条件
参数:
*	data  	map[string]interface{}	传入的查询数据
*	specs 	QuerySpec			查询规则
*	strict	bool				是否严格，非严格时，将data中未在specs出现的字段转为普通$eq条件
返回值:
*	bson.M	bson.M
*	error 	error
*/
func BuildQuery(data map[string]interface{}, specs QuerySpec, strict bool) (query bson.M, err error) {
	static, dynamic := remapQuerySpecByField(specs)

	var (
		exist                               bool
		staticResult, dynamicResult, result map[string]interface{} // 结果
		staticUsed, dynamicUsed, used       map[string]struct{}    // 记录使用过的数据
	)

	if staticResult, staticUsed, err = buildStaticQuery(data, static); err != nil {
		return nil, err
	}

	if dynamicResult, dynamicUsed, err = buildDynamicQuery(data, dynamic); err != nil {
		return nil, err
	}

	result = combine(staticResult, dynamicResult)
	used = combine(staticUsed, dynamicUsed)

	if !strict { // 如果不是严厉规则，将其余的数据加上，判断是否相等
		for k, v := range data {
			if _, exist = used[k]; !exist {
				result[k] = v
			}
		}
	}

	if len(result) == 0 {
		return nil, nil
	}

	query = clean(result)

	return query, nil
}

func combine[K string, V struct{} | interface{}](one, another map[K]V) map[K]V {
	if len(one) == 0 && len(another) == 0 {
		return nil
	}

	result := make(map[K]V, len(one)+len(another))

	for k := range one {
		result[k] = one[k]
	}

	for k := range another {
		result[k] = another[k]
	}

	return result
}

/*BuildQueryWithLogic 构建查询，并且支持逻辑操作
参数:
*	data  	map[string]interface{}		输入数据
*	specs 	QuerySpec					查询规则
*	strict	bool						是否严格(查看BuildQuery)
*	logic 	*LogicQuery					逻辑条件
返回值:
*	bson.M	bson.M						生成的查询条件
*	error 	error						可能的错误
*/
func BuildQueryWithLogic(data map[string]interface{}, specs QuerySpec, strict bool, logic *LogicQuery) (bson.M, error) {
	query, err := BuildQuery(data, specs, strict)
	if err != nil {
		return nil, err
	}

	return generate(logic, query), nil
}

// LogicQueryType 逻辑条件类型
type LogicQueryType int

const (
	// LogicNone 无逻辑条件，最后的数据
	LogicNone LogicQueryType = iota
	// LogicAnd	逻辑与
	LogicAnd
	// LogicOr	逻辑或
	LogicOr
)

// LogicQuery 逻辑查询条件
type LogicQuery struct {
	Type   LogicQueryType // 逻辑类型,如果为LogicOr时，Key不能为空,如果为其他类型时,Fields不能为空
	Fields []LogicQuery   // 内部成员
	Key    string         // 字段名
}

func generate(logic *LogicQuery, query bson.M) bson.M {
	switch logic.Type {
	case LogicAnd, LogicOr:
		return generateLogic(logic, query)
	case LogicNone:
		return query
	default:
		return query
	}
}

func generateLogic(logic *LogicQuery, query bson.M) bson.M {
	if logic.Type == LogicNone {
		return bson.M{logic.Key: query[logic.Key]}
	}

	fields := make([]bson.M, 0, len(logic.Fields))

	for i := range logic.Fields {
		fieldQuery := generateLogic(&logic.Fields[i], query)
		fields = append(fields, fieldQuery)
	}

	operator := "$and"

	if logic.Type == LogicOr {
		operator = "$or"
	}

	return bson.M{operator: fields}
}

/*clean 清理查询，主要是处理某个字段没有对应条件的情况，mongodb 会报错
参数:
*	query 	bson.M	原始条件
返回值:
*	bson.M	bson.M	清理后的条件
*/
func clean(query bson.M) bson.M {
	deleted := make([]string, 0, len(query))

	for k, v := range query {
		if value, ok := v.(map[string]interface{}); ok && len(value) == 0 { // 为了确保删除成功，不能在遍历中直接删除
			deleted = append(deleted, k)
		}
	}

	for _, key := range deleted {
		delete(query, key)
	}

	return query
}

const (
	defaultCapacity = 5
)

/*remapQuerySpecByField 通过查询的类型(静态、动态)重新组合查询条件
参数:
*	spec   	QuerySpec             	查询规则
返回值:
*	static 	map[string][]querySpec	静态规则
*	dynamic	map[string]Convert    	动态规则
逻辑:
1. 遍历每个查询条件
2. 根据静态/动态 区分
	1. 如果是动态，那么添加到动态组
	2. 如果是静态,根据适用的字段范围
		1. 如果是单个字段，那么直接设置
		2. 如果是多个字段，那么设置所有对应的字段
*/
func remapQuerySpecByField(spec QuerySpec) (static map[string][]*querySpec, dynamic map[string]Convert) {
	static = make(map[string][]*querySpec, defaultCapacity)

	staticField := func(fromField, forField, op string, convert Convert) {
		if _, exist := static[forField]; !exist {
			static[forField] = make([]*querySpec, 0, defaultCapacity)
		}

		static[forField] = append(static[forField], newQuerySpec(fromField, op, convert))
	}

	for k, v := range spec {
		if v.Dynamic { // 动态字段
			if len(dynamic) == 0 {
				dynamic = make(map[string]Convert, defaultCapacity)
			}

			dynamic[k] = v.Convert

			continue
		}

		// 静态字段
		if v.Field != "" { // 针对单个字段
			staticField(k, v.Field, v.Op, v.Convert)

			continue
		}

		for _, field := range v.Fields { // 针对多个字段，每个字段，都生效一次
			staticField(k, field, v.Op, v.Convert)
		}
	}

	return static, dynamic
}

type querySpec struct {
	key     string
	op      string
	convert Convert
}

func (q querySpec) GetValue(input string) (output interface{}, err error) {
	if q.convert == nil {
		return input, nil
	}

	return q.convert(input)
}

func newQuerySpec(key, op string, convert Convert) *querySpec {
	return &querySpec{
		key:     key,
		op:      op,
		convert: convert,
	}
}

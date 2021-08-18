[TOC]

# 概述

# 版本
|   版本  |   修改时间    |   修改人 |   修改内容 |
| --- | --- | --- | --- |
|   |    |  |   |
|  v0.2 | 2021-08-18 20:33:32 | 刘蕴唐 | 查看v0.2 变更记录  |
|   v0.1    |   2021-05-18 15:41:33 |  刘蕴唐    |   初步完成|


## v0.2 变更记录

1. 构建了Init() 方法，用于初始化，必须先调用一遍初始化
2. 更新了key的分隔符，从 . -> : ,方便reids中查看
3. 
# 详述


## 使用

### 初始化

**Init()**方法用户配置需要加载的文件路径和
```go
/*Init 初始化
参数:
*	path      	string                                    	配置文件路径
*	validators	map[string]govalidator.CustomTypeValidator	扩展的验证方法
返回值:
*/
func Init(path string, validators map[string]govalidator.CustomTypeValidator) 
```

### 添加

在**data/parameters.json**中添加一个json对象

```json
  {
    "key": "tron:chargeCheckInterval",
    "purpose": "查询充值结果间隔时间",
    "value": "30s",
    "description": "整数+单位,单位可以是s(秒),m(分钟),h(小时)",
    "validKey": "duration"
  }
```

|       字段|   类型|   含义|
|       ---|     ---|    --- |
|       key|    字符串| 参数key,格式必须为x:x,总长度在64 |
|       purpose|        字符串|  用途|
|       value|  字符串|  值|
|       description|    字符串| 值描述 |
|       validKey|       字符串| 验证key |


#### 验证key说明

可以使用**govalidator**的标签,或者自定义标签



####  自定义标签

1.  在**init.go**的**init()**中注册
2.  自定义的校验方法定义在valid.go中

```go
govalidator.CustomTypeTagMap.Set(`key`, keyValid)
```
如果是正则匹配，那么可以使用
```
regex(regexExpr *regexp.Regexp) func(i interface{}, o interface{}) bool
delimiter(content, delimiter string, count int) *regexp.Regexp
```

### 获取

使用接口**service.GetParameters()方法**获取
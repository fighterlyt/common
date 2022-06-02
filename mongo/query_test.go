package mongo

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/fighterlyt/log"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

func TestQuery(t *testing.T) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27018"))
	require.NoError(t, err)

	collection := client.Database("test").Collection("testquery")
	require.NoError(t, collection.Drop(context.Background()))

	times := 20
	for i := 0; i < times; i++ {
		data := TestData{
			A: i,
			B: times - i,
		}
		_, err := collection.InsertOne(context.Background(), data)
		require.NoError(t, err)
	}

	data := TestData{}
	result, count, err := Query(context.Background(), collection, NewQueryFilter(bson.M{"a": bson.M{"$lte": times}}, []string{"a"}, 5, 10, []string{"a"}, nil, data, false, nil))
	require.NoError(t, err)
	t.Log(count)

	for _, element := range result {
		t.Log(element)
	}
}

// BenchmarkQuery-8   	     200	   8724275 ns/op
func BenchmarkQuery(b *testing.B) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27018"))
	require.NoError(b, err)

	collection := client.Database("test").Collection("testquery")
	require.NoError(b, collection.Drop(context.Background()))

	times := 2000
	for i := 0; i < times; i++ {
		data := TestData{
			A: i,
			B: times - i,
		}
		_, err := collection.InsertOne(context.Background(), data)
		require.NoError(b, err)
	}

	data := TestData{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, count, _ := Query(context.Background(), collection, NewQueryFilter(bson.M{"a": bson.M{"$lte": times}}, []string{"a"}, 5, 1000, []string{"a"}, nil, data, true, nil))
		require.EqualValues(b, times, count)
	}
}

type TestData struct {
	A int `bson:"a"`
	B int `bson:"b"`
}

/*
   query_test.go:102: {a:1,attend:map[$gt:10 $lt:20],attendNo1:{$ne:15},attendNo2:{$ne:15},c:{$in:[]}}

*/
func TestBuildQuery(t *testing.T) {
	data := map[string]interface{}{
		"a":         1,
		"attend":    10,
		"attendMax": 20,
		"attendNo":  15,
		"c":         "a",
	}
	specs := QuerySpec{
		"attend": DBSpec{
			Op:    "$gt",
			Field: "attend",
		},
		"attendMax": DBSpec{
			Op:    "$lt",
			Field: "attend",
		},
		"attendNo": DBSpec{
			Op:     "$ne",
			Fields: []string{"attendNo1", "attendNo2"},
		},
		"c": {
			Op:    "$in",
			Field: "c",
			Convert: func(data string) (i interface{}, e error) {
				return []int{}, nil
			},
		},
	}
	query, err := BuildQuery(data, specs, false)
	require.NoError(t, err)
	t.Log(prettyBsonM(query))
}
func TestBuildQueryWithOr(t *testing.T) {
	data := map[string]interface{}{
		"a":         1,
		"attend":    10,
		"attendMax": 20,
		"attendNo":  15,
		"c":         "a",
	}
	specs := QuerySpec{
		"attend": DBSpec{
			Op:    "$gt",
			Field: "attend",
		},
		"attendMax": DBSpec{
			Op:    "$lt",
			Field: "attend",
		},
		"attendNo": DBSpec{
			Op:     "$ne",
			Fields: []string{"attendNo1", "attendNo2"},
		},
		"c": {
			Op:    "$in",
			Field: "c",
			Convert: func(data string) (i interface{}, e error) {
				return []int{}, nil
			},
		},
	}
	query, err := BuildQuery(data, specs, false)
	require.NoError(t, err)
	t.Log(query)
}

func TestBuildQueryWithDynamic(t *testing.T) {
	userID := "1"
	data := map[string]interface{}{
		"own":  "both",
		"kind": 1,
		"test": "1",
	}
	specs := QuerySpec{
		"own": {
			Dynamic: true,
			Convert: func(data string) (interface{}, error) {
				switch data {
				case "bid":
					return bson.M{"consumer.consumerID": userID}, nil
				case "create":
					return bson.M{"producer.producerID": userID}, nil
				case "both":
					return bson.M{"$or": []bson.M{{
						"consumer.consumerID": userID,
					}, {
						"producer.producerID": userID,
					}}}, nil
				default:
					return bson.M{}, fmt.Errorf("不支持的own参数[%s]", data)
				}
			},
		},
		"status": {
			Field: "status",
		},
		"kind": {
			Field: "kind",
		},
		"test": {
			Field: "producer.producerID",

			Op: "$in",
			Convert: func(data string) (interface{}, error) {
				return []string{"a"}, nil
			},
		},
	}
	//  map[$or:[map[consumer.consumerID:1] map[producer.producerID:1]] kind:map[:1] producer.producerID:map[$in:[a]]]
	// PASS
	query, err := BuildQuery(data, specs, true)
	require.NoError(t, err)
	t.Log(query)
}
func TestEnsureIndex(t *testing.T) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27018"))
	require.NoError(t, err)

	collection := client.Database("test").Collection("testIndex")
	indexes := []Index{
		{
			Name:    "test",
			Version: 9,
			Data: mongo.IndexModel{
				Keys:    bson.M{"a": bsonx.Int32(1), "b": bsonx.Int32(1)},
				Options: options.Index().SetUnique(true),
			},
		},
	}

	logger, _ := log.NewEasyLogger(false, false, ``, "test")

	require.NoError(t, EnsureIndex(collection, indexes, logger))
}

func TestZapBsonM(t *testing.T) {
	query := bson.M{"test": "a"}
	logger, _ := log.NewEasyLogger(false, false, ``, "test")

	logger.Debug("test", ZapBsonM("test", query))
}

func TestGenerate(t *testing.T) {
	data := map[string]interface{}{
		"a":         1,
		"attend":    10,
		"attendMax": 20,
		"attendNo":  15,
		"c":         "a",
	}
	specs := QuerySpec{
		"attend": DBSpec{
			Op:    "$gt",
			Field: "attend",
		},
		"attendMax": DBSpec{
			Op:    "$lt",
			Field: "attend",
		},
		"attendNo": DBSpec{
			Op:     "$ne",
			Fields: []string{"attendNo1", "attendNo2"},
		},
		"c": {
			Op:    "$in",
			Field: "c",
			Convert: func(data string) (i interface{}, e error) {
				return []int{}, nil
			},
		},
	}
	query, err := BuildQuery(data, specs, false)
	require.NoError(t, err)
	t.Log(prettyBsonM(query))

	logic := &LogicQuery{
		Type: LogicAnd,
		Fields: []LogicQuery{
			{
				Type: LogicOr,
				Fields: []LogicQuery{
					{
						Type: LogicNone,
						Key:  "attendNo1",
					},
					{
						Type: LogicNone,
						Key:  "attendNo2",
					},
				},
			},
			{
				Type: LogicAnd,
				Fields: []LogicQuery{
					{
						Type: LogicNone,
						Key:  "a",
					},
					{
						Type: LogicNone,
						Key:  "attend",
					},
				},
			},
		},
	}
	query = generate(logic, query)
	t.Log(prettyBsonM(query))
}

func TestBuildQuery2(t *testing.T) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27018"))
	require.NoError(t, err)

	collection := client.Database("test").Collection("testquery")

	type queryData struct {
		A int `bson:"a"`
		B int `bson:"b"`
		C int `bson:"c"`
		D int `bson:"d"`
	}

	data := map[string]interface{}{
		"a": "1",
		"b": "2",
		"c": "3",
		"d": "4",
		"5": "5",
	}
	specs := QuerySpec{
		"a": DBSpec{
			Op:      "$eq",
			Field:   "a",
			Convert: IntParse,
		},
		"b": DBSpec{
			Op:      "$eq",
			Field:   "b",
			Convert: IntParse,
		},
		"c": DBSpec{
			Op:      "$eq",
			Field:   "c",
			Convert: IntParse,
		},
		"d": DBSpec{
			Op:      "$eq",
			Field:   "d",
			Convert: IntParse,
		},
	}

	logic := &LogicQuery{
		Type: LogicOr,
		Fields: []LogicQuery{
			{
				Type: LogicAnd,
				Fields: []LogicQuery{
					{
						Type: LogicNone,
						Key:  "a",
					},
					{
						Type: LogicNone,
						Key:  "b",
					},
				},
			},
			{
				Type: LogicAnd,
				Fields: []LogicQuery{
					{
						Type: LogicNone,
						Key:  "c",
					},
					{
						Type: LogicNone,
						Key:  "d",
					},
				},
			},
		},
	}
	query, err := BuildQueryWithLogic(data, specs, true, logic)
	require.NoError(t, err)
	t.Log(prettyBsonM(query))

	result, count, err := Query(context.Background(), collection, NewQueryFilter(query, nil, 0, 0, nil, nil, &queryData{}, false, nil))
	require.NoError(t, err)
	t.Log(count)

	for _, element := range result {
		t.Log(element)
	}
}

func Test_convertSort(t *testing.T) {
	type args struct {
		sorts []string
	}
	tests := []struct {
		name string
		args args
		want bson.D
	}{
		{
			name: `增序倒序混合`,
			args: args{sorts: []string{"-a", "b", "-c", "d"}},
			want: bson.D{{"a", -1}, {"b", 1}, {"c", -1}, {"d", 1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertSort(tt.args.sorts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertSort() = %v, want %v", got, tt.want)
			}
		})
	}
}

package counter

import (
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	testCounter Counter
	value       = uint64(2)
	capacity    = int64(100)
	testMax     = int64(10)
)

func TestCounter_Set(t *testing.T) {
	testCounter = NewCounter(capacity)

	for i := int64(1); i <= testMax; i++ {
		testCounter.Set(i, value)
		require.EqualValuesf(t, i, testCounter.Count(), `第[%d]次添加后，数量应该是[%d]`, i, i)
	}

	for i := int64(1); i <= testMax; i++ {
		testCounter.Set(i, value)
		require.EqualValuesf(t, testMax, testCounter.Count(), `第二遍第[%d]次添加后，数量应该是[%d]`, i, testMax)
	}
}
func TestCounter_Clear(t *testing.T) {
	testCounter = NewCounter(capacity)

	for i := int64(1); i <= testMax; i++ {
		testCounter.Set(i, value)
		require.EqualValuesf(t, i, testCounter.Count(), `第[%d]次添加后，数量应该是[%d]`, i, i)
	}

	for i := int64(1); i <= testMax; i++ {
		testCounter.ClearIf(i, value+1)
		require.EqualValuesf(t, testMax, testCounter.Count(), `第[%d]次使用错误的值清理后，数量应该是[%d]`, i, testMax)
	}

	for i := int64(1); i <= testMax; i++ {
		testCounter.Clear(i)
		require.EqualValuesf(t, testMax-i, testCounter.Count(), `第[%d]次清理后，数量应该是[%d]`, i, testMax-i)
	}
}

func TestCounter_ClearIf(t *testing.T) {
	testCounter = NewCounter(capacity)

	for i := int64(1); i <= testMax; i++ {
		testCounter.Set(i, value)
		require.EqualValuesf(t, i, testCounter.Count(), `第[%d]次添加后，数量应该是[%d]`, i, i)
	}

	for i := int64(1); i <= testMax; i++ {
		testCounter.ClearIf(i, value)
		require.EqualValuesf(t, testMax-i, testCounter.Count(), `第[%d]次使用正确的值清理后，数量应该是[%d]`, i, testMax-i)
	}
}

func TestCounter_ClearAll(t *testing.T) {
	testCounter = NewCounter(capacity)

	for i := int64(1); i <= testMax; i++ {
		testCounter.Set(i, value)
		require.EqualValuesf(t, i, testCounter.Count(), `第[%d]次添加后，数量应该是[%d]`, i, i)
	}

	testCounter.ClearAll()

	require.EqualValuesf(t, 0, testCounter.Count(), `全部清理后正确的值清理后，数量应该是0`)
}

func TestCounter_ClearAllIfNot(t *testing.T) {
	testCounter = NewCounter(capacity)

	for i := int64(1); i <= testMax; i++ {
		if i%2 == 0 {
			testCounter.Set(i, value+1)
		} else {
			testCounter.Set(i, value)
		}

		require.EqualValuesf(t, i, testCounter.Count(), `第[%d]次添加后，数量应该是[%d]`, i, i)
	}

	testCounter.ClearAllIfNot(value + 1)

	require.EqualValuesf(t, testMax/2, testCounter.Count(), `全部清理后正确的值清理后，数量应该是%d`, testMax/2)
}

var (
	count int64 // 这个变量必须存在
)

//nolint:lll
/* BenchmarkCounter_Set 测试性能
cpu: Intel(R) Core(TM) i5-10600 CPU @ 3.30GHz
max/concurrent/times
1000,000/100/1,000                    BenchmarkCounter_Set-12    	      79	  15014068 ns/op	        81.86 总内存MB	    6678 B/op	     103 allocs/op
1000,000/100/10,000					  BenchmarkCounter_Set-12    	       7	 153010940 ns/op	        81.40 总内存MB	   50970 B/op	     213 allocs/op
1000,000/100/100,000                  BenchmarkCounter_Set-12    	       1	1529625665 ns/op	        43.05 总内存MB	  265048 B/op	     798 allocs/op
1000,000/100/100,0000                 BenchmarkCounter_Set-12    	       1	15577499179 ns/op	        43.01 总内存MB	  287768 B/op	    1012 allocs/op
1000,000/1,000/100,000                BenchmarkCounter_Set-12    	       1	19159734877 ns/op	        43.53 总内存MB	  822456 B/op	    3908 allocs/op
*/
func BenchmarkCounter_Set(b *testing.B) {
	testCounter = NewCounter(capacity * capacity * capacity)

	b.ReportAllocs()

	b.ResetTimer()

	max := int64(1000000)

	concurrent := 1000
	times := 100000

	r := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec

	randLock := &sync.Mutex{}

	int63n := func(data int64) int64 { // rand 不是病多发安全
		randLock.Lock()
		defer randLock.Unlock()

		return r.Int63n(data)
	}

	for k := 0; k < b.N; k++ {
		wg := &sync.WaitGroup{}
		wg.Add(concurrent)

		for i := 0; i < concurrent; i++ {
			go func() {
				for j := 0; j < times; j++ {
					if int63n(max)%2 == 0 { // 偶读，读取
						count = testCounter.Count()
					} else {
						testCounter.Set(int63n(max), value)
					}
				}

				wg.Done()
			}()
		}

		wg.Wait()

		m := &runtime.MemStats{}
		runtime.ReadMemStats(m)

		value := float64(m.Alloc)
		unit := `总内存字节`

		if m.Alloc > 1024*1024 {
			value /= (1024 * 1024)
			unit = `总内存MB`
		} else if m.Alloc > 1024 {
			value /= 1024
			unit = `总内存KB`
		}

		b.ReportMetric(value, unit)
	}
}

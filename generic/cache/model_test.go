package cache

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewCache(t *testing.T) {
	x := NewCache[int, *Data]()

	key := 10
	data := &Data{
		A: 10,
		B: "1",
	}

	// Add Get
	x.Add(key, data)

	require.Equal(t, data, x.Get(key))

	// Remove
	x.Remove(key)

	require.Nil(t, x.Get(key))

	// Update
	x.Add(key, data)

	newData := &Data{
		A: 11,
		B: "2",
	}

	x.Update(key, newData)

	require.Equal(t, newData, x.Get(key))
}

type Data struct {
	A int
	B string
}

func Benchmark_Cache_Get(b *testing.B) { // nolint:golint,revive
	x := NewCache[int, *Data]()

	times := 10000

	for i := 0; i < b.N; i++ {
		wg := sync.WaitGroup{}

		// 一个线程写，一个线程改，一个线程读,一个线程删除
		wg.Add(4)

		// 写入线程
		go func() {
			r := rand.New(rand.NewSource(time.Now().UnixNano())) // nolint:golint,gosec

			for i := 0; i < times; i++ {
				key := r.Intn(times)
				data := &Data{
					A: r.Intn(times),
					B: "aaa",
				}

				x.Add(key, data)
			}
			wg.Done()
		}()

		// 删除
		go func() {
			r := rand.New(rand.NewSource(time.Now().UnixNano())) // nolint:golint,gosec

			for i := 0; i < times; i++ {
				key := r.Intn(times)

				x.Remove(key)
			}
			wg.Done()
		}()

		// 修改
		go func() {
			r := rand.New(rand.NewSource(time.Now().UnixNano())) // nolint:golint,gosec

			for i := 0; i < times; i++ {
				key := r.Intn(times)
				data := &Data{
					A: r.Intn(times),
					B: "aaa",
				}

				x.Update(key, data)
			}
			wg.Done()
		}()

		// 获取
		go func() {
			r := rand.New(rand.NewSource(time.Now().UnixNano())) // nolint:golint,gosec

			for i := 0; i < times; i++ {
				key := r.Intn(times)

				x.Get(key)
			}
			wg.Done()
		}()

		wg.Wait()
	}
}

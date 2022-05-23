package cryptocurrency

import (
	"net"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/cpu"
)

func TestGetSymbol(t *testing.T) {
	type args struct {
		protocol Protocol
		test     bool
	}

	tests := []struct {
		name string
		args args
		want Symbol
	}{
		{
			name: "波场测试链",
			args: args{
				protocol: Trc20,
				test:     false,
			},
			want: USDT,
		},
		{
			name: `波场真实`,
			args: args{
				protocol: Trc20,
				test:     true,
			},
			want: SGMT,
		},
		{
			name: "以太坊测试链",
			args: args{
				protocol: Erc20,
				test:     false,
			},
			want: USDT,
		},
		{
			name: `以太坊真实`,
			args: args{
				protocol: Erc20,
				test:     true,
			},
			want: FLY,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSymbol(tt.args.protocol, tt.args.test)
			require.Equal(t, tt.want, got)
		})
	}
}

type NoPad struct {
	a uint64
	b uint64
	c uint64
}

func (np *NoPad) Increase() {
	atomic.AddUint64(&np.a, 1)
	atomic.AddUint64(&np.b, 1)
	atomic.AddUint64(&np.c, 1)
}

type Pad struct {
	a   uint64
	_p1 [8]uint64
	b   uint64
	_p2 [8]uint64
	c   uint64
	_p3 [8]uint64
}

func (p *Pad) Increase() {
	atomic.AddUint64(&p.a, 1)
	atomic.AddUint64(&p.b, 1)
	atomic.AddUint64(&p.c, 1)
}

func BenchmarkPad_Increase(b *testing.B) {
	pad := &Pad{}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pad.Increase()
		}
	})
}

func BenchmarkNoPad_Increase(b *testing.B) {
	nopad := &NoPad{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			nopad.Increase()
		}
	})
}

type Pad2 struct {
	_ cpu.CacheLinePad
	a uint64
	b uint64
	c uint64
	_ cpu.CacheLinePad
}

func (np *Pad2) Increase() {
	atomic.AddUint64(&np.a, 1)
	atomic.AddUint64(&np.b, 1)
	atomic.AddUint64(&np.c, 1)
}

func BenchmarkPad2_Increase(b *testing.B) {
	nopad := &Pad2{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			nopad.Increase()
		}
	})
}

var (
	matrixLength = 10000
)

func createMatrix(size int) [][]int64 {
	result := make([][]int64, size)

	for i := range result {
		result[i] = make([]int64, size)
	}

	return result
}
func BenchmarkMatrixCombination(b *testing.B) {
	matrixA := createMatrix(matrixLength)
	matrixB := createMatrix(matrixLength)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := 0; i < matrixLength; i++ {
			for j := 0; j < matrixLength; j++ {
				matrixA[i][j] += matrixB[i][j]
			}
		}
	}
}

func BenchmarkMatrixCombination2(b *testing.B) {
	matrixA := createMatrix(matrixLength)
	matrixB := createMatrix(matrixLength)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < matrixLength; i++ {
			for j := 0; j < matrixLength; j++ {
				matrixA[i][j] += matrixB[j][i]
			}
		}
	}
}

func BenchmarkMatrixReversedCombinationPerBlock(b *testing.B) {
	matrixA := createMatrix(matrixLength)
	matrixB := createMatrix(matrixLength)
	blockSize := 8

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < matrixLength; i += blockSize {
			for j := 0; j < matrixLength; j += blockSize {
				for ii := i; ii < i+blockSize; ii++ {
					for jj := j; jj < j+blockSize; jj++ {
						matrixA[ii][jj] += matrixB[ii][jj]
					}
				}
			}
		}
	}
}

type SimpleStruct struct {
	n int
}

const (
	M = 100000
)

type PaddedStruct struct {
	n int
	_ CacheLinePad
}

type CacheLinePad struct {
	_ [CacheLinePadSize]byte
}

const CacheLinePadSize = 64

// 同时给两个变量变化
func BenchmarkStructureFalseSharing(b *testing.B) {
	structA := SimpleStruct{}
	structB := SimpleStruct{}

	wg := sync.WaitGroup{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(2)

		go func() {
			for j := 0; j < M; j++ {
				structA.n += j
			}
			wg.Done()
		}()

		go func() {
			for j := 0; j < M; j++ {
				structB.n += j
			}
			wg.Done()
		}()

		wg.Wait()
	}
}

// 同时给两个变量变化
func BenchmarkStructureFalseSharing2(b *testing.B) {
	structA := PaddedStruct{}
	structB := PaddedStruct{}

	wg := sync.WaitGroup{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(2)

		go func() {
			for j := 0; j < M; j++ {
				structA.n += j
			}
			wg.Done()
		}()

		go func() {
			for j := 0; j < M; j++ {
				structB.n += j
			}
			wg.Done()
		}()

		wg.Wait()
	}
}

type NoCompact struct {
	A bool
	B int64
	C bool
}

type Compact struct {
	A bool
	C bool
	B int64
}

var (
	x     int64
	count = 100000
)

func BenchmarkNoCompact(b *testing.B) {
	data := make([]NoCompact, count)

	b.Log(unsafe.Sizeof(NoCompact{}))
	b.ResetTimer()

	j := int64(0)

	for i := 0; i < b.N; i++ {
		for _, elem := range data {
			j += elem.B
		}
	}

	x = j
}

func BenchmarkCompact(b *testing.B) {
	data := make([]Compact, count)

	b.Log(unsafe.Sizeof(Compact{}))

	b.ResetTimer()

	j := int64(0)

	for i := 0; i < b.N; i++ {
		for _, elem := range data {
			j += elem.B
		}
	}

	x = j
}

func TestListen(t *testing.T) {
	_, err := net.Listen(`tcp`, `:6060`)
	require.NoError(t, err)

	_, err = net.Listen(`tcp`, `:6060`)
	require.True(t, IsListenErrInUse(err))
}

func IsListenErrInUse(err error) bool {
	var (
		opErr        *net.OpError
		ok           bool
		syscallError *os.SyscallError
		errNo        syscall.Errno
	)

	if opErr, ok = err.(*net.OpError); !ok {
		return false
	}

	if syscallError, ok = opErr.Err.(*os.SyscallError); !ok {
		return false
	}

	if errNo, ok = syscallError.Err.(syscall.Errno); !ok {
		return false
	}

	return errNo == syscall.EADDRINUSE
}

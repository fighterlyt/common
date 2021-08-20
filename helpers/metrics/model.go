package metrics

import (
	"context"
	"os"
	"time"

	"github.com/fighterlyt/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
	"gitlab.com/nova_dubai/common/helpers"
	"gitlab.com/nova_dubai/common/model"
)

var (
	bg = context.Background()
)

type systemMetric interface {
	Run(ctx context.Context)
}
type System struct {
	cpu      *CPU
	mem      *MEM
	net      *Net
	p        *Process
	exit     chan struct{}
	ticker   *time.Ticker
	interval time.Duration
	shutdown model.Shutdown
	logger   log.Logger
}

func (s *System) Close() {
	s.shutdown.Close()
}

func (s *System) IsClosed() bool {
	return s.shutdown.IsClosed()
}

func (s *System) Key() string {
	return `system`
}

func (s *System) Name() string {
	return `系统`
}

func NewSystem(interval time.Duration, logger log.Logger) *System {
	pid := os.Getpid()

	return &System{
		cpu:      NewCPU(),
		mem:      NewMEM(),
		net:      NewNet(pid),
		p:        NewProcess(pid),
		ticker:   time.NewTicker(interval),
		exit:     make(chan struct{}, 0),
		interval: interval,
		shutdown: model.NewShutdown(),
		logger:   logger,
	}
}

func (s *System) Start() {
	helpers.EnsureGo(s.logger, func() {
		for {
			select {
			case <-s.ticker.C:
				s.Run()
			case <-s.exit:
				return
			}
		}
	})
}

func (s *System) Run() {
	ctx, cancel := context.WithTimeout(bg, s.interval)
	defer cancel()

	helpers.EnsureGo(s.logger, func() {
		s.cpu.Run(ctx)
	})

	helpers.EnsureGo(s.logger, func() {
		s.mem.Run(ctx)
	})

	helpers.EnsureGo(s.logger, func() {
		s.net.Run(ctx)
	})

	helpers.EnsureGo(s.logger, func() {
		s.p.Run(ctx)
	})
}

func (s *System) Finish() {
	close(s.exit)
}

type CPU struct {
	counts      prometheus.Gauge
	countsValue int
	// percentage      prometheus.Gauge
	// percentageValue float64
	err error
}

func NewCPU() *CPU {
	return &CPU{
		counts: promauto.NewGauge(prometheus.GaugeOpts{
			Name: `system:cpu:counts`,
			Help: `CPU数量(包括逻辑核心)`,
		}),
		// percentage: prometheus.NewGauge(prometheus.GaugeOpts{
		// 	Name: `system:cpu:percentage`,
		// 	Help: `CPU使用率`,
		// }),
	}
}

func (c *CPU) Run(ctx context.Context) {
	if c.countsValue != 0 {
		return
	}

	if c.countsValue, c.err = cpu.CountsWithContext(ctx, true); c.err == nil {
		c.counts.Set(float64(c.countsValue))
	}
}

type MEM struct {
	totalValue       uint64
	total            prometheus.Gauge
	usedPercentValue float64
	usedPercent      prometheus.Gauge
	data             *mem.VirtualMemoryStat
	err              error
}

func NewMEM() *MEM {
	return &MEM{
		total: promauto.NewGauge(prometheus.GaugeOpts{
			Name: `system:mem:total`,
			Help: `内存数量(字节)`,
		}),
		usedPercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: `system:mem:usedPercent`,
			Help: `使用率`,
		}),
	}
}

func (m *MEM) Run(ctx context.Context) {
	if m.data, m.err = mem.VirtualMemoryWithContext(ctx); m.err == nil {
		m.totalValue = m.data.Total
		m.usedPercentValue = m.data.UsedPercent

		m.total.Set(float64(m.totalValue))
		m.usedPercent.Set(m.usedPercentValue)
	}
}

type Net struct {
	countsValue int
	counts      prometheus.Gauge
	data        []net.ConnectionStat
	err         error
	pid         int
}

func NewNet(pid int) *Net {
	return &Net{
		counts: promauto.NewGauge(prometheus.GaugeOpts{
			Name: `system:net:counts`,
			Help: `网络连接数`,
		}),
		pid: pid,
	}
}

func (n *Net) Run(ctx context.Context) {
	if n.data, n.err = net.ConnectionsPidWithContext(ctx, `tcp4`, int32(n.pid)); n.err == nil {
		n.countsValue = 0
		for _, elem := range n.data {
			if elem.Status == `ESTABLISHED` {
				n.countsValue++
			}
		}

		n.counts.Set(float64(n.countsValue))
		return
	}
}

type Process struct {
	pid             int
	memPercentValue float32
	memPercent      prometheus.Gauge
	numFDValue      int32                   // 使用的文件描述符数量
	numFD           prometheus.Gauge        //
	openFilesValue  []process.OpenFilesStat // 打开文件数
	openFiles       prometheus.Gauge

	rlimit *prometheus.GaugeVec

	err error

	rlimits []process.RlimitStat
	p       *process.Process
}

func NewProcess(pid int) *Process {
	return &Process{pid: pid,
		memPercent: promauto.NewGauge(prometheus.GaugeOpts{
			Name: `system:process:memPercent`,
			Help: `进程内存使用率`,
		}),
		numFD: promauto.NewGauge(prometheus.GaugeOpts{
			Name: `system:process:numFD`,
			Help: `使用文件描述符数量`,
		}),
		openFiles: promauto.NewGauge(prometheus.GaugeOpts{
			Name: `system:process:openFiles`,
			Help: `打开文件数`,
		}),
		rlimit: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: `system:process:rlimit`,
			Help: `rlimit`,
		}, []string{`kind`}),
	}
}

func (p *Process) Run(ctx context.Context) {
	if p.p, p.err = process.NewProcess(int32(p.pid)); p.err != nil {
		return
	}

	if p.rlimits, p.err = p.p.RlimitUsageWithContext(ctx, true); p.err == nil {
		if len(p.rlimits) == 16 { // 16项

			p.rlimit.WithLabelValues(`soft`).Set(float64(p.rlimits[7].Soft))
			p.rlimit.WithLabelValues(`hard`).Set(float64(p.rlimits[7].Hard))
			p.rlimit.WithLabelValues(`used`).Set(float64(p.rlimits[7].Used))
		}
	}

	if p.numFDValue, p.err = p.p.NumFDsWithContext(ctx); p.err == nil {
		p.numFD.Set(float64(p.numFDValue))
	}

	if p.openFilesValue, p.err = p.p.OpenFilesWithContext(ctx); p.err == nil {
		p.openFiles.Set(float64(len(p.openFilesValue)))
	}

	if p.memPercentValue, p.err = p.p.MemoryPercentWithContext(ctx); p.err == nil {
		p.memPercent.Set(float64(p.memPercentValue))
	}
}

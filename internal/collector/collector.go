// Package collector exposes the Bbox metrics as a prometheus.Collector.
package collector

import (
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"bbox7_exporter/internal/bbox"
)

type Collector struct {
	client *bbox.Client
	logger *slog.Logger
}

func New(client *bbox.Client, logger *slog.Logger) *Collector {
	return &Collector{client: client, logger: logger}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range allDescs {
		ch <- d
	}
}

// Collect authenticates, then collects each endpoint independently — one
// failure doesn't block the others.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	if err := c.client.Login(); err != nil {
		gauge(ch, upDesc, 0)
		c.logger.Error("authentication failed", "err", err)
		return
	}
	gauge(ch, upDesc, 1)

	c.collectDevice(ch)
	c.collectCPU(ch)
	c.collectMem(ch)
	c.collectWAN(ch)
	c.collectLAN(ch)
	c.collectWiFi(ch)
	c.collectHosts(ch)
	c.collectServices(ch)
	c.collectVoip(ch)
}

// --- helpers ------------------------------------------------------------------

func gauge(ch chan<- prometheus.Metric, d *prometheus.Desc, v float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(d, prometheus.GaugeValue, v, labels...)
}

func counter(ch chan<- prometheus.Metric, d *prometheus.Desc, v float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(d, prometheus.CounterValue, v, labels...)
}

func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// ep emits bbox_scrape_endpoint_success and logs any failure.
func (c *Collector) ep(ch chan<- prometheus.Metric, name string, err error) bool {
	gauge(ch, scrapeDesc, b2f(err == nil), name)
	if err != nil {
		c.logger.Warn("endpoint failed", "endpoint", name, "err", err)
		return false
	}
	return true
}

// --- sections -----------------------------------------------------------------

func (c *Collector) collectDevice(ch chan<- prometheus.Metric) {
	d, err := c.client.Device()
	if !c.ep(ch, "device", err) {
		return
	}
	gauge(ch, deviceInfoDesc, 1, d.ModelName, d.SerialNumber, d.Main.Version)
	gauge(ch, deviceStatusDesc, float64(d.Status))
	gauge(ch, uptimeDesc, float64(d.Uptime))
	counter(ch, bootsDesc, float64(d.NumberOfBoots))
	for k, v := range d.Using {
		gauge(ch, faiDesc, float64(v), k)
	}
}

func (c *Collector) collectCPU(ch chan<- prometheus.Metric) {
	cpu, err := c.client.CPU()
	if !c.ep(ch, "device/cpu", err) {
		return
	}
	for mode, jiffies := range cpu.Time {
		if mode == "total" { // sum of the other modes: avoids double counting
			continue
		}
		counter(ch, cpuTimeDesc, jiffies/100.0, mode) // USER_HZ = 100
	}
	gauge(ch, procDesc, cpu.Process.Running, "running")
	gauge(ch, procDesc, cpu.Process.Blocked, "blocked")
	counter(ch, procCreatedDesc, cpu.Process.Created)
	gauge(ch, tempDesc, cpu.Temperature.Main/1000.0) // milli-°C → °C
}

func (c *Collector) collectMem(ch chan<- prometheus.Metric) {
	m, err := c.client.Mem()
	if !c.ep(ch, "device/mem", err) {
		return
	}
	gauge(ch, memDesc, m.Total, "total")
	gauge(ch, memDesc, m.Free, "free")
	gauge(ch, memDesc, m.Cached, "cached")
	gauge(ch, memDesc, m.CommittedAs, "committed")
}

func (c *Collector) collectWAN(ch chan<- prometheus.Metric) {
	if ip, err := c.client.WanIP(); c.ep(ch, "wan/ip", err) {
		gauge(ch, wanInternetDesc, float64(ip.Internet.State))
		gauge(ch, wanIPUpDesc, b2f(ip.IP.State == "Up"))
		gauge(ch, wanIP6UpDesc, b2f(ip.IP.IP6State == "Up"))
	}
	if s, err := c.client.WanStats(); c.ep(ch, "wan/ip/stats", err) {
		emit := func(dir string, x bbox.Counter) {
			counter(ch, wanBytesDesc, x.Bytes, dir)
			counter(ch, wanPacketsDesc, x.Packets, dir)
			counter(ch, wanErrorsDesc, x.PacketsErrors, dir)
			counter(ch, wanDiscardsDesc, x.PacketsDiscards, dir)
			gauge(ch, wanBandwidthDesc, x.Bandwidth, dir)
			gauge(ch, wanMaxBwDesc, x.MaxBandwidth, dir)
			gauge(ch, wanContractBwDesc, x.ContractualBandwidth, dir)
			gauge(ch, wanOccupDesc, x.Occupation, dir)
		}
		emit("rx", s.Rx)
		emit("tx", s.Tx)
	}
}

func (c *Collector) collectLAN(ch chan<- prometheus.Metric) {
	l, err := c.client.LanStats()
	if !c.ep(ch, "lan/stats", err) {
		return
	}
	counter(ch, lanBytesDesc, l.Rx.Bytes, "rx")
	counter(ch, lanBytesDesc, l.Tx.Bytes, "tx")
	counter(ch, lanPacketsDesc, l.Rx.Packets, "rx")
	counter(ch, lanPacketsDesc, l.Tx.Packets, "tx")
	for _, p := range l.Port {
		port := strconv.Itoa(p.Index)
		counter(ch, lanPortBytesDesc, p.Rx.Bytes, port, "rx")
		counter(ch, lanPortBytesDesc, p.Tx.Bytes, port, "tx")
		counter(ch, lanPortPacketsDesc, p.Rx.Packets, port, "rx")
		counter(ch, lanPortPacketsDesc, p.Tx.Packets, port, "tx")
		gauge(ch, lanPortBwDesc, p.Rx.Bandwidth, port, "rx")
		gauge(ch, lanPortBwDesc, p.Tx.Bandwidth, port, "tx")
	}
}

func (c *Collector) collectWiFi(ch chan<- prometheus.Metric) {
	bands := []struct{ api, label string }{
		{"24", "2.4"},
		{"5", "5"},
		{"6", "6"},
	}
	for _, b := range bands {
		if s, err := c.client.WirelessStats(b.api); c.ep(ch, "wireless/"+b.api+"/stats", err) {
			emit := func(dir string, x bbox.Counter) {
				counter(ch, wifiBytesDesc, x.Bytes, b.label, dir)
				counter(ch, wifiPacketsDesc, x.Packets, b.label, dir)
				counter(ch, wifiErrorsDesc, x.PacketsErrors, b.label, dir)
				counter(ch, wifiDiscardsDesc, x.PacketsDiscards, b.label, dir)
			}
			emit("rx", s.Rx)
			emit("tx", s.Tx)
		}
		if r, err := c.client.WirelessRadio(b.api); c.ep(ch, "wireless/"+b.api, err) {
			gauge(ch, wifiRadioDesc, b2f(r.Enable == 1), b.label)
			gauge(ch, wifiChannelDesc, float64(r.CurrentChannel), b.label)
			gauge(ch, wifiBwDesc, float64(r.CurrentBandwidth), b.label)
			if r.Standard != "" {
				gauge(ch, wifiStdDesc, 1, b.label, r.Standard)
			}
		}
	}
}

func (c *Collector) collectHosts(ch chan<- prometheus.Metric) {
	list, err := c.client.Hosts()
	if !c.ep(ch, "hosts", err) {
		return
	}
	counts := map[string]int{}
	total := 0
	for _, h := range list {
		if h.Active == 1 {
			counts[h.Link]++
			total++
		}
	}
	for link, n := range counts {
		gauge(ch, hostsConnDesc, float64(n), link)
	}
	gauge(ch, hostsTotalDesc, float64(total))
}

func (c *Collector) collectServices(ch chan<- prometheus.Metric) {
	s, err := c.client.Services()
	if !c.ep(ch, "services", err) {
		return
	}
	for _, x := range []struct {
		name           string
		status, enable int
	}{
		{"firewall", s.Firewall.Status, s.Firewall.Enable},
		{"dhcp", s.Dhcp.Status, s.Dhcp.Enable},
		{"nat", s.Nat.Status, s.Nat.Enable},
		{"gamermode", s.Gamermode.Status, s.Gamermode.Enable},
		{"hotspot", s.Hotspot.Status, s.Hotspot.Enable},
		{"dyndns", s.Dyndns.State, s.Dyndns.Enable},
	} {
		gauge(ch, serviceStatusDesc, float64(x.status), x.name)
		gauge(ch, serviceEnabledDesc, float64(x.enable), x.name)
	}
}

func (c *Collector) collectVoip(ch chan<- prometheus.Metric) {
	lines, err := c.client.Voip()
	if !c.ep(ch, "voip", err) {
		return
	}
	for _, l := range lines {
		id := strconv.Itoa(l.ID)
		gauge(ch, voipUpDesc, b2f(l.Status == "Up"), id, l.URI)
		gauge(ch, voipCallDesc, 1, id, l.CallState)
	}
}

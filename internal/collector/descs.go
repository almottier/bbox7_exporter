package collector

import "github.com/prometheus/client_golang/prometheus"

const namespace = "bbox"

func desc(name, help string, labels ...string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, "", name), help, labels, nil)
}

var (
	upDesc     = desc("up", "1 if the last authentication against the Bbox succeeded.")
	scrapeDesc = desc("scrape_endpoint_success", "1 if the endpoint was collected successfully.", "endpoint")

	// device
	deviceInfoDesc   = desc("device_info", "Static box information (constant value 1).", "model_name", "serial", "firmware")
	deviceStatusDesc = desc("device_status", "Box status.")
	uptimeDesc       = desc("device_uptime_seconds", "Box uptime in seconds.")
	bootsDesc        = desc("device_boots_total", "Number of reboots since commissioning.")
	faiDesc          = desc("device_fai_active", "Active ISP connection by type (1=active).", "using")

	// cpu / mem / temperature
	cpuTimeDesc     = desc("cpu_time_seconds_total", "Cumulative CPU time by mode (seconds).", "mode")
	procDesc        = desc("processes", "Number of processes by state.", "state")
	procCreatedDesc = desc("processes_created_total", "Number of processes created since boot.")
	tempDesc        = desc("temperature_celsius", "Internal box temperature (°C).")
	memDesc         = desc("memory_kbytes", "Memory by type (kilobytes).", "type")

	// wan
	wanInternetDesc   = desc("wan_internet_state", "Internet connection state (2=up).")
	wanIPUpDesc       = desc("wan_ip_up", "1 if the WAN IPv4 is Up.")
	wanIP6UpDesc      = desc("wan_ip6_up", "1 if the WAN IPv6 is Up.")
	wanBytesDesc      = desc("wan_bytes_total", "Cumulative WAN bytes.", "direction")
	wanPacketsDesc    = desc("wan_packets_total", "Cumulative WAN packets.", "direction")
	wanErrorsDesc     = desc("wan_packets_errors_total", "WAN packet errors.", "direction")
	wanDiscardsDesc   = desc("wan_packets_discards_total", "Discarded WAN packets.", "direction")
	wanBandwidthDesc  = desc("wan_bandwidth_kbits", "Instantaneous WAN throughput (kbit/s).", "direction")
	wanMaxBwDesc      = desc("wan_max_bandwidth_kbits", "Maximum negotiated WAN throughput (kbit/s).", "direction")
	wanContractBwDesc = desc("wan_contractual_bandwidth_kbits", "Contractual WAN throughput (kbit/s).", "direction")
	wanOccupDesc      = desc("wan_occupation_percent", "Instantaneous WAN link occupation (%).", "direction")

	// lan
	lanBytesDesc       = desc("lan_bytes_total", "Cumulative LAN bytes (global).", "direction")
	lanPacketsDesc     = desc("lan_packets_total", "Cumulative LAN packets (global).", "direction")
	lanPortBytesDesc   = desc("lan_port_bytes_total", "Cumulative bytes per LAN port.", "port", "direction")
	lanPortPacketsDesc = desc("lan_port_packets_total", "Cumulative packets per LAN port.", "port", "direction")
	lanPortBwDesc      = desc("lan_port_bandwidth_kbits", "Instantaneous throughput per LAN port (kbit/s).", "port", "direction")

	// wifi (3 bands: 2.4 / 5 / 6 GHz)
	wifiBytesDesc    = desc("wifi_bytes_total", "Cumulative Wi-Fi bytes per band.", "band", "direction")
	wifiPacketsDesc  = desc("wifi_packets_total", "Cumulative Wi-Fi packets per band.", "band", "direction")
	wifiErrorsDesc   = desc("wifi_packets_errors_total", "Wi-Fi packet errors per band.", "band", "direction")
	wifiDiscardsDesc = desc("wifi_packets_discards_total", "Discarded Wi-Fi packets per band.", "band", "direction")
	wifiRadioDesc    = desc("wifi_radio_enabled", "1 if the band's radio is enabled.", "band")
	wifiChannelDesc  = desc("wifi_channel", "Current channel of the band.", "band")
	wifiBwDesc       = desc("wifi_bandwidth_mhz", "Current channel width (MHz).", "band")
	wifiStdDesc      = desc("wifi_standard_info", "Wi-Fi standard of the band (constant value 1).", "band", "standard")

	// hosts
	hostsConnDesc  = desc("hosts_connected", "Number of active devices per link type.", "link")
	hostsTotalDesc = desc("hosts_total", "Total number of active devices.")

	// services
	serviceStatusDesc  = desc("service_status", "Status of a service (codes depend on the box).", "service")
	serviceEnabledDesc = desc("service_enabled", "1 if the service is enabled.", "service")

	// voip
	voipUpDesc   = desc("voip_line_up", "1 if the VoIP line is Up.", "id", "uri")
	voipCallDesc = desc("voip_line_callstate_info", "Call state of the VoIP line (constant value 1).", "id", "callstate")
)

// allDescs backs Describe.
var allDescs = []*prometheus.Desc{
	upDesc, scrapeDesc,
	deviceInfoDesc, deviceStatusDesc, uptimeDesc, bootsDesc, faiDesc,
	cpuTimeDesc, procDesc, procCreatedDesc, tempDesc, memDesc,
	wanInternetDesc, wanIPUpDesc, wanIP6UpDesc, wanBytesDesc, wanPacketsDesc,
	wanErrorsDesc, wanDiscardsDesc, wanBandwidthDesc, wanMaxBwDesc, wanContractBwDesc, wanOccupDesc,
	lanBytesDesc, lanPacketsDesc, lanPortBytesDesc, lanPortPacketsDesc, lanPortBwDesc,
	wifiBytesDesc, wifiPacketsDesc, wifiErrorsDesc, wifiDiscardsDesc,
	wifiRadioDesc, wifiChannelDesc, wifiBwDesc, wifiStdDesc,
	hostsConnDesc, hostsTotalDesc,
	serviceStatusDesc, serviceEnabledDesc,
	voipUpDesc, voipCallDesc,
}

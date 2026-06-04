// Package bbox is a minimal client for the local /api/v1 API of the Bbox
// Wi-Fi 7 (F@st5696b). All responses are single-element arrays:
//
//	[ { "<root>": { ... } } ]
package bbox

import (
	"encoding/json"
	"strconv"
	"strings"
)

// jsonNum decodes a scalar that may be a number (123) or a string ("123") —
// the firmware quotes some counters inconsistently.
type jsonNum float64

func (n *jsonNum) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*n = 0
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*n = jsonNum(f)
	return nil
}

// Counter holds the rx/tx counters from the stats endpoints; absent fields stay zero.
type Counter struct {
	Packets              float64 `json:"packets"`
	Bytes                float64 `json:"bytes"`
	PacketsErrors        float64 `json:"packetserrors"`
	PacketsDiscards      float64 `json:"packetsdiscards"`
	Occupation           float64 `json:"occupation"`
	Bandwidth            float64 `json:"bandwidth"`
	MaxBandwidth         float64 `json:"maxBandwidth"`
	ContractualBandwidth float64 `json:"contractualBandwidth"`
}

// UnmarshalJSON tolerates fields the firmware returns as strings (see jsonNum).
func (c *Counter) UnmarshalJSON(b []byte) error {
	var a struct {
		Packets              jsonNum `json:"packets"`
		Bytes                jsonNum `json:"bytes"`
		PacketsErrors        jsonNum `json:"packetserrors"`
		PacketsDiscards      jsonNum `json:"packetsdiscards"`
		Occupation           jsonNum `json:"occupation"`
		Bandwidth            jsonNum `json:"bandwidth"`
		MaxBandwidth         jsonNum `json:"maxBandwidth"`
		ContractualBandwidth jsonNum `json:"contractualBandwidth"`
	}
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*c = Counter{
		Packets:              float64(a.Packets),
		Bytes:                float64(a.Bytes),
		PacketsErrors:        float64(a.PacketsErrors),
		PacketsDiscards:      float64(a.PacketsDiscards),
		Occupation:           float64(a.Occupation),
		Bandwidth:            float64(a.Bandwidth),
		MaxBandwidth:         float64(a.MaxBandwidth),
		ContractualBandwidth: float64(a.ContractualBandwidth),
	}
	return nil
}

// Stats: rx/tx pair (WAN, Wi-Fi).
type Stats struct {
	Rx Counter `json:"rx"`
	Tx Counter `json:"tx"`
}

// Device: /device
type Device struct {
	Status        int    `json:"status"`
	NumberOfBoots int    `json:"numberofboots"`
	ModelName     string `json:"modelname"`
	SerialNumber  string `json:"serialnumber"`
	Main          struct {
		Version string `json:"version"`
	} `json:"main"`
	Uptime int64          `json:"uptime"`
	Using  map[string]int `json:"using"` // ftth, ipv4, ipv6, adsl, vdsl
}

// CPU: /device/cpu
type CPU struct {
	Time    map[string]float64 `json:"time"` // total, user, nice, system, io, idle, irq (jiffies)
	Process struct {
		Created float64 `json:"created"`
		Running float64 `json:"running"`
		Blocked float64 `json:"blocked"`
	} `json:"process"`
	Temperature struct {
		Main float64 `json:"main"` // milli-degrees Celsius
	} `json:"temperature"`
}

// Mem: /device/mem (kilobytes)
type Mem struct {
	Total       float64 `json:"total"`
	Free        float64 `json:"free"`
	Cached      float64 `json:"cached"`
	CommittedAs float64 `json:"committedas"`
}

// WanIP: /wan/ip
type WanIP struct {
	Internet struct {
		State int `json:"state"`
	} `json:"internet"`
	IP struct {
		State    string `json:"state"`    // "Up"/"Down"
		IP6State string `json:"ip6state"` // "Up"/"Down"
	} `json:"ip"`
}

// LanPort: a physical port in /lan/stats.
type LanPort struct {
	Index int     `json:"index"`
	Rx    Counter `json:"rx"`
	Tx    Counter `json:"tx"`
}

// LanStats: /lan/stats
type LanStats struct {
	Rx   Counter   `json:"rx"`
	Tx   Counter   `json:"tx"`
	Port []LanPort `json:"port"`
}

// WifiRadio: /wireless/{band} (radio sub-object)
type WifiRadio struct {
	Enable           int    `json:"enable"`
	State            int    `json:"state"`
	Standard         string `json:"standard"`
	CurrentChannel   int    `json:"current_channel"`
	CurrentBandwidth int    `json:"current_bandwidth"` // MHz
}

// Host: a device in /hosts.
type Host struct {
	Active int    `json:"active"`
	Link   string `json:"link"` // "Ethernet", "Wifi 5", ...
}

// ServiceState: generic status of a service in /services.
type ServiceState struct {
	Status  int `json:"status"`
	Enable  int `json:"enable"`
	NbRules int `json:"nbrules"`
}

// Services: /services (subset we use).
type Services struct {
	Firewall  ServiceState `json:"firewall"`
	Dhcp      ServiceState `json:"dhcp"`
	Nat       ServiceState `json:"nat"`
	Gamermode ServiceState `json:"gamermode"`
	Hotspot   ServiceState `json:"hotspot"`
	Dyndns    struct {
		State  int `json:"state"`
		Enable int `json:"enable"`
	} `json:"dyndns"`
}

// VoipLine: a line in /voip.
type VoipLine struct {
	ID        int    `json:"id"`
	Status    string `json:"status"`    // "Up"
	CallState string `json:"callstate"` // "Idle", ...
	URI       string `json:"uri"`
}

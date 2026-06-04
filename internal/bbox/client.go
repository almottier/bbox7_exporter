package bbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// lockoutDuration: the box locks auth after repeated failures; self-applied to
// avoid worsening a lockout.
const lockoutDuration = 120 * time.Second

type Client struct {
	baseURL  string
	password string
	http     *http.Client

	mu          sync.Mutex
	cookies     []*http.Cookie
	lockedUntil time.Time
}

// New creates a client. endpoint must be https; TLS is never disabled.
//
// resolveIP is optional: when set, TCP dials go to that IP while TLS/SNI still
// use the endpoint host — the "curl --resolve" trick, no /etc/hosts needed.
func New(endpoint, password, resolveIP string) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" {
		return nil, fmt.Errorf("invalid endpoint (https required): %q", endpoint)
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	if resolveIP != "" {
		host := u.Hostname()
		dialer := &net.Dialer{Timeout: 10 * time.Second}
		httpClient.Transport = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// only redirect the endpoint host; keep the port (SNI stays the hostname)
				if h, port, err := net.SplitHostPort(addr); err == nil && h == host {
					addr = net.JoinHostPort(resolveIP, port)
				}
				return dialer.DialContext(ctx, network, addr)
			},
		}
	}

	return &Client{
		baseURL:  strings.TrimRight(u.String(), "/") + "/api/v1",
		password: password,
		http:     httpClient,
	}, nil
}

func (c *Client) Login() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().Before(c.lockedUntil) {
		return fmt.Errorf("authentication skipped until %s (anti-lockout %s)",
			c.lockedUntil.Format(time.RFC3339), lockoutDuration)
	}

	resp, err := c.http.PostForm(c.baseURL+"/login", url.Values{"password": {c.password}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		c.lockedUntil = time.Now().Add(lockoutDuration)
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return fmt.Errorf("login HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	cookies := resp.Cookies()
	if len(cookies) == 0 {
		return fmt.Errorf("login: no session cookie returned")
	}
	c.cookies = cookies
	return nil
}

func (c *Client) get(path string, v any) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cache-Control", "no-cache")

	c.mu.Lock()
	for _, ck := range c.cookies {
		req.AddCookie(ck)
	}
	c.mu.Unlock()

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: HTTP %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// fetch returns the single element of the array response.
func fetch[T any](c *Client, path string) (T, error) {
	var arr []T
	var zero T
	if err := c.get(path, &arr); err != nil {
		return zero, err
	}
	if len(arr) == 0 {
		return zero, fmt.Errorf("empty response for %s", path)
	}
	return arr[0], nil
}

// --- typed access per endpoint ------------------------------------------------

func (c *Client) Device() (Device, error) {
	w, err := fetch[struct {
		Device Device `json:"device"`
	}](c, "/device")
	return w.Device, err
}

func (c *Client) CPU() (CPU, error) {
	w, err := fetch[struct {
		Device struct {
			CPU CPU `json:"cpu"`
		} `json:"device"`
	}](c, "/device/cpu")
	return w.Device.CPU, err
}

func (c *Client) Mem() (Mem, error) {
	w, err := fetch[struct {
		Device struct {
			Mem Mem `json:"mem"`
		} `json:"device"`
	}](c, "/device/mem")
	return w.Device.Mem, err
}

func (c *Client) WanIP() (WanIP, error) {
	w, err := fetch[struct {
		Wan WanIP `json:"wan"`
	}](c, "/wan/ip")
	return w.Wan, err
}

func (c *Client) WanStats() (Stats, error) {
	w, err := fetch[struct {
		Wan struct {
			IP struct {
				Stats Stats `json:"stats"`
			} `json:"ip"`
		} `json:"wan"`
	}](c, "/wan/ip/stats")
	return w.Wan.IP.Stats, err
}

func (c *Client) LanStats() (LanStats, error) {
	w, err := fetch[struct {
		Lan struct {
			Stats LanStats `json:"stats"`
		} `json:"lan"`
	}](c, "/lan/stats")
	return w.Lan.Stats, err
}

// WirelessStats returns the rx/tx counters for a band ("24", "5", "6").
func (c *Client) WirelessStats(band string) (Stats, error) {
	w, err := fetch[struct {
		Wireless struct {
			SSID struct {
				Stats Stats `json:"stats"`
			} `json:"ssid"`
		} `json:"wireless"`
	}](c, "/wireless/"+band+"/stats")
	return w.Wireless.SSID.Stats, err
}

// WirelessRadio returns the radio state of a band ("24", "5", "6").
func (c *Client) WirelessRadio(band string) (WifiRadio, error) {
	w, err := fetch[struct {
		Wireless struct {
			Radio WifiRadio `json:"radio"`
		} `json:"wireless"`
	}](c, "/wireless/"+band)
	return w.Wireless.Radio, err
}

func (c *Client) Hosts() ([]Host, error) {
	w, err := fetch[struct {
		Hosts struct {
			List []Host `json:"list"`
		} `json:"hosts"`
	}](c, "/hosts")
	return w.Hosts.List, err
}

func (c *Client) Services() (Services, error) {
	w, err := fetch[struct {
		Services Services `json:"services"`
	}](c, "/services")
	return w.Services, err
}

func (c *Client) Voip() ([]VoipLine, error) {
	w, err := fetch[struct {
		Voip []VoipLine `json:"voip"`
	}](c, "/voip")
	return w.Voip, err
}

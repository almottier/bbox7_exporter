// bbox7_exporter: Prometheus exporter dedicated to the Bbox Bouygues Telecom
// Wi-Fi 7 (model F@st5696b).
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"bbox7_exporter/internal/bbox"
	"bbox7_exporter/internal/collector"
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	endpoint := flag.String("endpoint", env("BBOX_EXPORTER_ENDPOINT", "https://mabbox.bytel.fr"),
		"Base URL of the Bbox.")
	password := flag.String("password", os.Getenv("BBOX_EXPORTER_PASSWORD"),
		"Admin password (prefer the BBOX_EXPORTER_PASSWORD environment variable).")
	resolve := flag.String("resolve", os.Getenv("BBOX_EXPORTER_RESOLVE"),
		"Box IP to dial directly, bypassing DNS/etc-hosts (TLS still verified against the endpoint host).")
	addr := flag.String("web.listen-address", env("BBOX_EXPORTER_ADDR", ":9311"),
		"HTTP server listen address.")
	path := flag.String("web.telemetry-path", env("BBOX_EXPORTER_METRICS_PATH", "/metrics"),
		"Path under which to expose metrics.")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	if *password == "" {
		logger.Error("password required (--password or BBOX_EXPORTER_PASSWORD)")
		os.Exit(1)
	}

	client, err := bbox.New(*endpoint, *password, *resolve)
	if err != nil {
		logger.Error("invalid client", "err", err)
		os.Exit(1)
	}
	prometheus.MustRegister(collector.New(client, logger))

	http.Handle(*path, promhttp.Handler())
	http.HandleFunc("/-/healthy", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprintln(w, "OK") })
	http.HandleFunc("/-/ready", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprintln(w, "OK") })
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, `<html><head><title>Bbox7 Exporter</title></head>`+
			`<body><h1>Bbox7 Exporter</h1><p><a href="%s">Metrics</a></p></body></html>`, *path)
	})

	logger.Info("starting bbox7_exporter", "addr", *addr, "endpoint", *endpoint)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		logger.Error("HTTP server stopped", "err", err)
		os.Exit(1)
	}
}

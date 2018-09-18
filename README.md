# Container Exporter

Lightweight Prometheus exporter for Linux containers/cgroups.

Container Exporter collects metrics from within a Linux container.
Useful to scrape container metrics without read access to host
metrics. If you have host access or at least to its metrics, you
should consider [cadvisor](https://github.com/google/cadvisor)
instead.

# Usage

Copy, `chmod +x` and run
[`container-exporter`](https://github.com/jcmoraisjr/container-exporter/releases)
as a daemon before start the main process. The default listening
port is `9009`, change the default port with the optional `-b :8000`
command-line argument.

# Metrics

The following metrics are available:

|Name|Description|
|---|---|
|`container_memory_usage_bytes`|Current memory usage in bytes|
|`container_cpu_usage_seconds_total`|Cumulative CPU time consumed in seconds|
|`container_network_transmit_bytes_total`|Cumulative count of bytes transmitted|
|`container_network_receive_bytes_total`|Cumulative count of bytes received|

CPU and network metrics are counters and might be used with
[`rate()`](https://prometheus.io/docs/prometheus/latest/querying/functions/#rate())
to calculate per-second average rate.

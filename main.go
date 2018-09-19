/*
Copyright 2018 Joao Morais.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
)

type prometheusCollector struct {
	metrics []*prometheusMetric
}

type prometheusMetric struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
	valueCalc []sysReader
}

type sysReader interface {
	ReadNumber() float64
	ReadLabelValues() []string
}

type sysNumeric struct {
	fileName    string
	labelValues []string
	divisor     int
}

func (c *prometheusCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric.desc
	}
}

func (c *prometheusCollector) Collect(ch chan<- prometheus.Metric) {
	for _, metric := range c.metrics {
		for _, calc := range metric.valueCalc {
			ch <- prometheus.MustNewConstMetric(
				metric.desc,
				metric.valueType,
				calc.ReadNumber(),
				calc.ReadLabelValues()...)
		}
	}
}

func (n *sysNumeric) ReadNumber() float64 {
	b, _ := ioutil.ReadFile(n.fileName)
	f, _ := strconv.ParseFloat(strings.TrimSpace(string(b)), 64)
	if n.divisor > 0 {
		f = f / float64(n.divisor)
	}
	return f
}

func (n *sysNumeric) ReadLabelValues() []string {
	return n.labelValues
}

func networkStatsTxRxBytes() ([]sysReader, []sysReader) {
	var networkLabels []string
	intfs, _ := ioutil.ReadDir("/sys/class/net")
	for _, intf := range intfs {
		networkLabels = append(networkLabels, intf.Name())
	}
	netStat := func(io string) []sysReader {
		var calc []sysReader
		for _, netLabel := range networkLabels {
			calc = append(calc, &sysNumeric{
				fileName:    fmt.Sprintf("/sys/class/net/%s/statistics/%s_bytes", netLabel, io),
				labelValues: []string{netLabel},
			})
		}
		return calc
	}
	return netStat("tx"), netStat("rx")
}

func newRegistry() *prometheus.Registry {
	netStatBytesTx, netStatBytesRx := networkStatsTxRxBytes()

	metrics := []*prometheusMetric{
		{
			desc:      prometheus.NewDesc("container_memory_usage_bytes", "Current memory usage in bytes", []string{}, nil),
			valueType: prometheus.GaugeValue,
			valueCalc: []sysReader{
				&sysNumeric{fileName: "/sys/fs/cgroup/memory/memory.usage_in_bytes"},
			},
		}, {
			desc:      prometheus.NewDesc("container_cpu_usage_seconds_total", "Cumulative cpu time consumed in seconds", []string{}, nil),
			valueType: prometheus.CounterValue,
			valueCalc: []sysReader{
				&sysNumeric{fileName: "/sys/fs/cgroup/cpu,cpuacct/cpuacct.usage", divisor: 1e9},
			},
		}, {
			desc:      prometheus.NewDesc("container_network_transmit_bytes_total", "Cumulative count of bytes transmitted", []string{"interface"}, nil),
			valueType: prometheus.CounterValue,
			valueCalc: netStatBytesTx,
		}, {
			desc:      prometheus.NewDesc("container_network_receive_bytes_total", "Cumulative count of bytes received", []string{"interface"}, nil),
			valueType: prometheus.CounterValue,
			valueCalc: netStatBytesRx,
		},
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(&prometheusCollector{
		metrics: metrics,
	})
	return reg
}

var (
	flags    = pflag.NewFlagSet("container-exporter", pflag.ExitOnError)
	flagAddr = flags.StringP("bind-address", "b", ":9009", "IP address and port number to serve.")
)

func main() {
	flags.Parse(os.Args)
	http.Handle("/", promhttp.HandlerFor(newRegistry(), promhttp.HandlerOpts{}))
	if err := http.ListenAndServe(*flagAddr, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

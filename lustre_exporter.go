// (C) Copyright 2017 Hewlett Packard Enterprise Development LP
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	
	"github.com/joehandzik/lustre_exporter/exporter"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var (
	showVersion	= flag.Bool("version", false, "Print version information.")
	listenAddress	= flag.String("web.listen-address", ":9169", "Address to use to expose Lustre metrics.")
	metricsPath	= flag.String("web.telemetry-path", "/metrics", "Path to use to expose Lustre metrics.")
)

func init() {
	prometheus.MustRegister(version.NewCollector("lustre_exporter"))
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("lustre_exporter"))
		os.Exit(0)
	}

	log.Infoln("Starting lustre_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	prometheus.MustRegister(lustre.NewExporter())

	handler := prometheus.Handler()

	http.Handle(*metricsPath, handler)
	http.HandleFunc("/" func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Lustre Exporter</title></head>
			<body>
			<h1>Lustre Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Infoln("Listening on", *listenAddress)
	err = http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
} 

# Lustre Metrics Exporter

[![Go Report Card](https://goreportcard.com/badge/github.com/HewlettPackard/lustre_exporter)](https://goreportcard.com/report/github.com/HewlettPackard/lustre_exporter)
[![Build Status](https://travis-ci.org/HewlettPackard/lustre_exporter.svg?branch=master)](https://travis-ci.org/HewlettPackard/lustre_exporter)

[Prometheus](https://prometheus.io/) exporter for Lustre metrics.

## Getting

```
go get github.com/HewlettPackard/lustre_exporter
```

## Building


```
cd $GOPATH/src/github.com/HewlettPackard/lustre_exporter
make
```

## Running

```
./lustre_exporter <flags>
```

### Flags

Boolean (True/False)

* collector.ost - Enable OST metrics
* collector.mdt - Enable MDT metrics
* collector.mgs - Enable MGS metrics
* collector.mds - Enable MDS metrics
* collector.client - Enable client metrics
* collector.generic - Enable generic metrics
* collector.lnet - Enable lnet metrics

## What's exported?

All Lustre procfs and procsys data from all nodes running the Lustre Exporter that we perceive as valuable data is exported or can be added to be exported (we don't have any known major gaps that anyone cares about, so if you see something missing, please file an issue!).

See the issues tab for all known issues. This project is actively maintained by HPE, so you should see a reasonably quick response if you identify a gap.

## Contributing

To contribute to this HPE project, you'll need to fill out a CLA (Contributor License Agreement). If you would like to contribute anything more than a bug fix (feature, architectural change, etc), please file an issue and we'll get in touch with you to have you fill out the CLA. 

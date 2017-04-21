# Lustre Metrics Exporter

Prometheus exporter for Lustre metrics.

## Getting

go get github.com/HewlettPackard/lustre_exporter

## Building

cd $GOPATH/src/github.com/HewlettPackard/lustre_exporter

make

## Running

./lustre_exporter

### Flags

TODO, but the current plan is to create flags that would define what node metrics are pulled from a running instance (e.g., OSS and MDS metrics would each have their own flag to disable). Also, we'd have flags to disable non-procfs metrics.

### What's exported?

Design plans

1. Export all proc data from all nodes running the Lustre Exporter that can function as a counter type (will save histogram-type work for later).
  - STATUS: Mostly complete as far as we can tell. We still want to parse rpcstats and jobstats at minimum before calling this complete.
2. Identify redundant data (if it exists).
  - Deduplication would be done, at first, by enabling flags to identify the node type with a configuration flag.
  - STATUS: No problems so far, saving for future work.
3. Add in:
  - Histogram data
    - STATUS: We have some of this via the 'brw_stats' file. We have created the histograms within Grafana in early tests.
  - Other data sources (CLI data that isn't present in /proc, for example). Users will be able to disable non-proc sources via a configuration flag.
    - STATUS: Not yet started
    
## Contributing

To contribute to this HPE project, you'll need to fill out a CLA (Contributor License Agreement). If you would like to contribute anything more than a bug fix (feature, architectural change, etc), please file an issue and we'll get in touch with you to have you fill out the CLA. 

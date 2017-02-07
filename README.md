# Lustre Metrics Exporter

Prometheus exporter for Lustre metrics.

## Building and running

TODO

### Flags

TODO

### What's exported?

Design plans

1. Export all proc data from all nodes running the Lustre Exporter that can function as a counter type (will save histogram-type work for later).
2. Identify redundant data (if it exists).
- Deduplication would be done, at first, by enabling flags to identify the node type with a configuration flag.
3. Add in:
- Histogram data
- Other data sources (CLI data that isn't present in /proc, for example). Users will be able to disable non-proc sources via a configuration flag.

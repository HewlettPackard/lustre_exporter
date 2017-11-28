# Change Log

## [v1.1.0](https://github.com/HewlettPackard/lustre_exporter/tree/v1.1.0) (2017-11-28)
[Full Changelog](https://github.com/HewlettPackard/lustre_exporter/compare/v1.0.0...v1.1.0)

**Closed issues:**

- Replace govendor with dep [\#127](https://github.com/HewlettPackard/lustre_exporter/issues/127)

**Merged pull requests:**

- Updating to use dep instead of govendor [\#128](https://github.com/HewlettPackard/lustre_exporter/pull/128) ([joehandzik](https://github.com/joehandzik))

## [v1.0.0](https://github.com/HewlettPackard/lustre_exporter/tree/v1.0.0) (2017-11-16)
[Full Changelog](https://github.com/HewlettPackard/lustre_exporter/compare/v1.0.0-rc.1...v1.0.0)

**Implemented enhancements:**

- Introduce "Core" and "Extended" as flag concepts [\#116](https://github.com/HewlettPackard/lustre_exporter/issues/116)

**Closed issues:**

- Feature request: Switch to kingpin [\#120](https://github.com/HewlettPackard/lustre_exporter/issues/120)
- Feature request: -collector Options [\#119](https://github.com/HewlettPackard/lustre_exporter/issues/119)
- client: label dimensions inconsistent with previously collected metrics in the same metric family [\#100](https://github.com/HewlettPackard/lustre_exporter/issues/100)

**Merged pull requests:**

- Update changelog to prepare for version 1.0.0 tag [\#126](https://github.com/HewlettPackard/lustre_exporter/pull/126) ([joehandzik](https://github.com/joehandzik))
- Add systemd directory for service files [\#124](https://github.com/HewlettPackard/lustre_exporter/pull/124) ([joehandzik](https://github.com/joehandzik))
- Clean up simple lint error [\#123](https://github.com/HewlettPackard/lustre_exporter/pull/123) ([joehandzik](https://github.com/joehandzik))
- Add double-dash to example flags [\#122](https://github.com/HewlettPackard/lustre_exporter/pull/122) ([joehandzik](https://github.com/joehandzik))
- Support Kingpin and improve help text [\#121](https://github.com/HewlettPackard/lustre_exporter/pull/121) ([joehandzik](https://github.com/joehandzik))
- Labeling overhaul [\#118](https://github.com/HewlettPackard/lustre_exporter/pull/118) ([joehandzik](https://github.com/joehandzik))
- Change flag behavior to introduce metric priority levels [\#117](https://github.com/HewlettPackard/lustre_exporter/pull/117) ([joehandzik](https://github.com/joehandzik))
- Update version and CHANGELOG [\#115](https://github.com/HewlettPackard/lustre_exporter/pull/115) ([joehandzik](https://github.com/joehandzik))

## [v1.0.0-rc.1](https://github.com/HewlettPackard/lustre_exporter/tree/v1.0.0-rc.1) (2017-07-31)
[Full Changelog](https://github.com/HewlettPackard/lustre_exporter/compare/v0.01.0...v1.0.0-rc.1)

**Implemented enhancements:**

- Ignored command line flags? [\#109](https://github.com/HewlettPackard/lustre_exporter/issues/109)

**Closed issues:**

- Convert internal representations to floating point values [\#111](https://github.com/HewlettPackard/lustre_exporter/issues/111)
- RPC and console metrics names  [\#96](https://github.com/HewlettPackard/lustre_exporter/issues/96)
- Cases exist where type label is unpopulated for some client stats [\#94](https://github.com/HewlettPackard/lustre_exporter/issues/94)
- Test interface FS [\#93](https://github.com/HewlettPackard/lustre_exporter/issues/93)
- Argument ordering inconsistent between stats and job\_stats methods [\#89](https://github.com/HewlettPackard/lustre_exporter/issues/89)
- Make stats and job\_stats patterns the same [\#88](https://github.com/HewlettPackard/lustre_exporter/issues/88)
- Change kilobytes\_\* metrics to \*\_kilobytes [\#76](https://github.com/HewlettPackard/lustre_exporter/issues/76)
- inode metrics [\#72](https://github.com/HewlettPackard/lustre_exporter/issues/72)
- Add lnet statistics [\#70](https://github.com/HewlettPackard/lustre_exporter/issues/70)
- Add client statistics [\#69](https://github.com/HewlettPackard/lustre_exporter/issues/69)
- /proc/fs/lustre/sptlrpc/encrypt\_page\_pools stats [\#68](https://github.com/HewlettPackard/lustre_exporter/issues/68)
- Missing ldlm pool stats? [\#67](https://github.com/HewlettPackard/lustre_exporter/issues/67)
- Converge md\_stats and some job\_stats statistics into a single metric [\#64](https://github.com/HewlettPackard/lustre_exporter/issues/64)
- Missing md\_stats from code [\#62](https://github.com/HewlettPackard/lustre_exporter/issues/62)
- `build user` not parsing correctly [\#56](https://github.com/HewlettPackard/lustre_exporter/issues/56)
- Semantic Versioning and CHANGELOG.md [\#55](https://github.com/HewlettPackard/lustre_exporter/issues/55)
- Normalize size labels [\#50](https://github.com/HewlettPackard/lustre_exporter/issues/50)
- Lowercase labels for metrics [\#49](https://github.com/HewlettPackard/lustre_exporter/issues/49)
- Fix help text for obdfilter/\*/brw\_size [\#48](https://github.com/HewlettPackard/lustre_exporter/issues/48)
- Convert counters to gauges where applicable [\#47](https://github.com/HewlettPackard/lustre_exporter/issues/47)
- /proc/fs/lustre/health\_check [\#46](https://github.com/HewlettPackard/lustre_exporter/issues/46)
- Bring metric names in compliance with prometheus standards [\#16](https://github.com/HewlettPackard/lustre_exporter/issues/16)

**Merged pull requests:**

- Update CHANGELOG [\#114](https://github.com/HewlettPackard/lustre_exporter/pull/114) ([joehandzik](https://github.com/joehandzik))
- Log to stdout what sub-components are enabled [\#113](https://github.com/HewlettPackard/lustre_exporter/pull/113) ([joehandzik](https://github.com/joehandzik))
- Use float64 instead of uint64 for all Prom metrics [\#112](https://github.com/HewlettPackard/lustre_exporter/pull/112) ([joehandzik](https://github.com/joehandzik))
- Replace 'max wait time' with 'out of mem' [\#108](https://github.com/HewlettPackard/lustre_exporter/pull/108) ([roclark](https://github.com/roclark))
- Update README [\#106](https://github.com/HewlettPackard/lustre_exporter/pull/106) ([joehandzik](https://github.com/joehandzik))
- Add Kerberos stats [\#105](https://github.com/HewlettPackard/lustre_exporter/pull/105) ([roclark](https://github.com/roclark))
- Move all number parsing to separate function [\#103](https://github.com/HewlettPackard/lustre_exporter/pull/103) ([roclark](https://github.com/roclark))
- Refactor Testing Infrastructre [\#102](https://github.com/HewlettPackard/lustre_exporter/pull/102) ([roclark](https://github.com/roclark))
- Update write\_minimum\_size\_bytes help text [\#101](https://github.com/HewlettPackard/lustre_exporter/pull/101) ([roclark](https://github.com/roclark))
- Remove support for errors that won't happen [\#99](https://github.com/HewlettPackard/lustre_exporter/pull/99) ([joehandzik](https://github.com/joehandzik))
- Add wildcard to client metrics [\#98](https://github.com/HewlettPackard/lustre_exporter/pull/98) ([roclark](https://github.com/roclark))
- Fix RPC metric names and types [\#97](https://github.com/HewlettPackard/lustre_exporter/pull/97) ([roclark](https://github.com/roclark))
- Don't populate empty labels for rpc\_stats [\#95](https://github.com/HewlettPackard/lustre_exporter/pull/95) ([joehandzik](https://github.com/joehandzik))
- Add `gometalinter` support [\#92](https://github.com/HewlettPackard/lustre_exporter/pull/92) ([mjtrangoni](https://github.com/mjtrangoni))
- Reorder getStats\*Metrics arguments [\#91](https://github.com/HewlettPackard/lustre_exporter/pull/91) ([joehandzik](https://github.com/joehandzik))
- Add RPC stats for the client [\#90](https://github.com/HewlettPackard/lustre_exporter/pull/90) ([roclark](https://github.com/roclark))
- Coalesce some statistics into a single metric [\#87](https://github.com/HewlettPackard/lustre_exporter/pull/87) ([joehandzik](https://github.com/joehandzik))
- Convert brw\_stats sizes to bytes [\#86](https://github.com/HewlettPackard/lustre_exporter/pull/86) ([joehandzik](https://github.com/joehandzik))
- Cleanup errors caught by megacheck [\#85](https://github.com/HewlettPackard/lustre_exporter/pull/85) ([joehandzik](https://github.com/joehandzik))
- Add Megacheck [\#84](https://github.com/HewlettPackard/lustre_exporter/pull/84) ([mjtrangoni](https://github.com/mjtrangoni))
- Update float parsing patterns [\#83](https://github.com/HewlettPackard/lustre_exporter/pull/83) ([roclark](https://github.com/roclark))
- Add LNET statistics [\#82](https://github.com/HewlettPackard/lustre_exporter/pull/82) ([roclark](https://github.com/roclark))
- Allow multiple sources [\#81](https://github.com/HewlettPackard/lustre_exporter/pull/81) ([roclark](https://github.com/roclark))
- Add CHANGELOG.md [\#80](https://github.com/HewlettPackard/lustre_exporter/pull/80) ([joehandzik](https://github.com/joehandzik))
- Add missing ldlm pool stats [\#79](https://github.com/HewlettPackard/lustre_exporter/pull/79) ([roclark](https://github.com/roclark))
- Expand metrics collected by stats files [\#78](https://github.com/HewlettPackard/lustre_exporter/pull/78) ([roclark](https://github.com/roclark))
- Change kilobytes\_\* to \*\_kilobytes [\#77](https://github.com/HewlettPackard/lustre_exporter/pull/77) ([joehandzik](https://github.com/joehandzik))
- Add client statistics [\#75](https://github.com/HewlettPackard/lustre_exporter/pull/75) ([roclark](https://github.com/roclark))
- Add unit tests [\#74](https://github.com/HewlettPackard/lustre_exporter/pull/74) ([roclark](https://github.com/roclark))
- Tweak metric naming for file\_\* and \*\_now values [\#73](https://github.com/HewlettPackard/lustre_exporter/pull/73) ([joehandzik](https://github.com/joehandzik))
- Skip empty jobstats [\#71](https://github.com/HewlettPackard/lustre_exporter/pull/71) ([roclark](https://github.com/roclark))
- Add missing md\_stats [\#66](https://github.com/HewlettPackard/lustre_exporter/pull/66) ([joehandzik](https://github.com/joehandzik))
- Fix range variable c captured by func literal [\#65](https://github.com/HewlettPackard/lustre_exporter/pull/65) ([mjtrangoni](https://github.com/mjtrangoni))
- Update version to comply with semantic versioning [\#63](https://github.com/HewlettPackard/lustre_exporter/pull/63) ([joehandzik](https://github.com/joehandzik))
- Makefile improvements [\#61](https://github.com/HewlettPackard/lustre_exporter/pull/61) ([mjtrangoni](https://github.com/mjtrangoni))
- Skip jobstats metric if they don't exist [\#60](https://github.com/HewlettPackard/lustre_exporter/pull/60) ([roclark](https://github.com/roclark))
- Add Travis-CI [\#59](https://github.com/HewlettPackard/lustre_exporter/pull/59) ([roclark](https://github.com/roclark))
- Properly parse BuildUser [\#58](https://github.com/HewlettPackard/lustre_exporter/pull/58) ([roclark](https://github.com/roclark))
- Specify Prometheus metric type for all metrics [\#57](https://github.com/HewlettPackard/lustre_exporter/pull/57) ([roclark](https://github.com/roclark))
- Consolidate CounterValue functions [\#54](https://github.com/HewlettPackard/lustre_exporter/pull/54) ([roclark](https://github.com/roclark))
- Add 'health\_check' metric [\#53](https://github.com/HewlettPackard/lustre_exporter/pull/53) ([roclark](https://github.com/roclark))
- Update brw\_size help text [\#52](https://github.com/HewlettPackard/lustre_exporter/pull/52) ([joehandzik](https://github.com/joehandzik))
- Lowercase all metric labels [\#51](https://github.com/HewlettPackard/lustre_exporter/pull/51) ([joehandzik](https://github.com/joehandzik))
- Make metric names comply with Prometheus standards [\#44](https://github.com/HewlettPackard/lustre_exporter/pull/44) ([joehandzik](https://github.com/joehandzik))
- Update version to 1.00.0 RC1 [\#43](https://github.com/HewlettPackard/lustre_exporter/pull/43) ([joehandzik](https://github.com/joehandzik))

## [v0.01.0](https://github.com/HewlettPackard/lustre_exporter/tree/v0.01.0) (2017-05-03)
**Closed issues:**

- Add data from mdt/\* [\#27](https://github.com/HewlettPackard/lustre_exporter/issues/27)
- Correct label values [\#25](https://github.com/HewlettPackard/lustre_exporter/issues/25)
- Refactor metric enumeration [\#24](https://github.com/HewlettPackard/lustre_exporter/issues/24)
- All metrics trailing a 'stats' file are missing [\#19](https://github.com/HewlettPackard/lustre_exporter/issues/19)
- Ensure that all existing code handles cases of missing data with graceful partial failure  [\#17](https://github.com/HewlettPackard/lustre_exporter/issues/17)
- Metrics with 'lustre\_lustre\_" prefix [\#11](https://github.com/HewlettPackard/lustre_exporter/issues/11)

**Merged pull requests:**

- Remove '-alpha' from version [\#42](https://github.com/HewlettPackard/lustre_exporter/pull/42) ([joehandzik](https://github.com/joehandzik))
- Fix syntax error in initialization [\#41](https://github.com/HewlettPackard/lustre_exporter/pull/41) ([roclark](https://github.com/roclark))
- Slight change to some jobstat names to remove \_id [\#40](https://github.com/HewlettPackard/lustre_exporter/pull/40) ([joehandzik](https://github.com/joehandzik))
- Include metrics from md\_stats [\#39](https://github.com/HewlettPackard/lustre_exporter/pull/39) ([roclark](https://github.com/roclark))
- Add all jobstats metrics [\#38](https://github.com/HewlettPackard/lustre_exporter/pull/38) ([roclark](https://github.com/roclark))
- Allow custom names [\#37](https://github.com/HewlettPackard/lustre_exporter/pull/37) ([roclark](https://github.com/roclark))
- Refactor metric initialization to use structs [\#36](https://github.com/HewlettPackard/lustre_exporter/pull/36) ([roclark](https://github.com/roclark))
- Reduce redundant code [\#35](https://github.com/HewlettPackard/lustre_exporter/pull/35) ([roclark](https://github.com/roclark))
- Update README \(various items\) [\#34](https://github.com/HewlettPackard/lustre_exporter/pull/34) ([joehandzik](https://github.com/joehandzik))
- Refactor stats to use structs [\#33](https://github.com/HewlettPackard/lustre_exporter/pull/33) ([roclark](https://github.com/roclark))
- Use 'gofmt -s' to simplify code [\#32](https://github.com/HewlettPackard/lustre_exporter/pull/32) ([joehandzik](https://github.com/joehandzik))
- Clean up golint warnings [\#31](https://github.com/HewlettPackard/lustre_exporter/pull/31) ([joehandzik](https://github.com/joehandzik))
- Clean up spelling mistakes [\#30](https://github.com/HewlettPackard/lustre_exporter/pull/30) ([joehandzik](https://github.com/joehandzik))
- Add support for MDT metrics [\#29](https://github.com/HewlettPackard/lustre_exporter/pull/29) ([joehandzik](https://github.com/joehandzik))
- Fix incorrect naming convention for OSTs [\#28](https://github.com/HewlettPackard/lustre_exporter/pull/28) ([joehandzik](https://github.com/joehandzik))
- Collect Job Stats [\#26](https://github.com/HewlettPackard/lustre_exporter/pull/26) ([roclark](https://github.com/roclark))
- Vendor the lustre\_exporter [\#23](https://github.com/HewlettPackard/lustre_exporter/pull/23) ([joehandzik](https://github.com/joehandzik))
- Add Contributing details to README [\#22](https://github.com/HewlettPackard/lustre_exporter/pull/22) ([joehandzik](https://github.com/joehandzik))
- Add distributed lock manager singlestat files [\#21](https://github.com/HewlettPackard/lustre_exporter/pull/21) ([joehandzik](https://github.com/joehandzik))
- Reset metricType to single [\#20](https://github.com/HewlettPackard/lustre_exporter/pull/20) ([joehandzik](https://github.com/joehandzik))
- Add 'make clean' support to Makefile [\#18](https://github.com/HewlettPackard/lustre_exporter/pull/18) ([joehandzik](https://github.com/joehandzik))
- Update repo-dependent names to HewlettPackard [\#15](https://github.com/HewlettPackard/lustre_exporter/pull/15) ([joehandzik](https://github.com/joehandzik))
- Update README.md [\#14](https://github.com/HewlettPackard/lustre_exporter/pull/14) ([joehandzik](https://github.com/joehandzik))
- Fix nil map error [\#13](https://github.com/HewlettPackard/lustre_exporter/pull/13) ([roclark](https://github.com/roclark))
- Remove redundant metric prefixes [\#12](https://github.com/HewlettPackard/lustre_exporter/pull/12) ([roclark](https://github.com/roclark))
- Flatten 'stats' map and simplify 'stats' logic [\#10](https://github.com/HewlettPackard/lustre_exporter/pull/10) ([roclark](https://github.com/roclark))
- Parse BRW Stats [\#9](https://github.com/HewlettPackard/lustre_exporter/pull/9) ([roclark](https://github.com/roclark))
- Add MDS metrics [\#8](https://github.com/HewlettPackard/lustre_exporter/pull/8) ([roclark](https://github.com/roclark))
- Remove broken metric [\#7](https://github.com/HewlettPackard/lustre_exporter/pull/7) ([roclark](https://github.com/roclark))
- Parse metrics from stats file [\#6](https://github.com/HewlettPackard/lustre_exporter/pull/6) ([roclark](https://github.com/roclark))
- Add MGS metrics [\#5](https://github.com/HewlettPackard/lustre_exporter/pull/5) ([roclark](https://github.com/roclark))
- Add metrics from the obdfilter [\#4](https://github.com/HewlettPackard/lustre_exporter/pull/4) ([roclark](https://github.com/roclark))
- Add help text [\#3](https://github.com/HewlettPackard/lustre_exporter/pull/3) ([roclark](https://github.com/roclark))
- Skip metrics that don't exist [\#2](https://github.com/HewlettPackard/lustre_exporter/pull/2) ([roclark](https://github.com/roclark))
- sources/procfs: Add the first metric [\#1](https://github.com/HewlettPackard/lustre_exporter/pull/1) ([joehandzik](https://github.com/joehandzik))



\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*
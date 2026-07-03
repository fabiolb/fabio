# Changelog

## [v1.7.2](https://github.com/fabiolb/fabio/tree/v1.7.2) (2026-07-03)

[Full Changelog](https://github.com/fabiolb/fabio/compare/v1.7.1...v1.7.2)

**Fixed bugs:**

- Fix Benchmark tests. [\#1041](https://github.com/fabiolb/fabio/pull/1041) ([tristanmorgan](https://github.com/tristanmorgan))

**Closed issues:**

- URL query contains semicolon, which is no longer a supported separator; parts of the query may be stripped when parsed; see golang.org/issue/25192 [\#955](https://github.com/fabiolb/fabio/issues/955)
- Datadog tracing [\#772](https://github.com/fabiolb/fabio/issues/772)

**Merged pull requests:**

- Bump github.com/tg123/go-htpasswd from 1.2.4 to 1.2.5 [\#1048](https://github.com/fabiolb/fabio/pull/1048) ([dependabot[bot]](https://github.com/apps/dependabot))
- Bump golang.org/x/net from 0.55.0 to 0.56.0 [\#1047](https://github.com/fabiolb/fabio/pull/1047) ([dependabot[bot]](https://github.com/apps/dependabot))
- Drop in replacement for armon/go-proxyproto [\#1045](https://github.com/fabiolb/fabio/pull/1045) ([tristanmorgan](https://github.com/tristanmorgan))
- Bump github.com/hashicorp/consul/api from 1.34.2 to 1.34.3 [\#1044](https://github.com/fabiolb/fabio/pull/1044) ([dependabot[bot]](https://github.com/apps/dependabot))
- Bump golang.org/x/net from 0.53.0 to 0.55.0 [\#1043](https://github.com/fabiolb/fabio/pull/1043) ([dependabot[bot]](https://github.com/apps/dependabot))
- Bump google.golang.org/grpc from 1.80.0 to 1.81.1 [\#1040](https://github.com/fabiolb/fabio/pull/1040) ([dependabot[bot]](https://github.com/apps/dependabot))
- Update http://github.com/mwitkow/grpc-proxy to the latest version [\#1039](https://github.com/fabiolb/fabio/pull/1039) ([tristanmorgan](https://github.com/tristanmorgan))

## [v1.7.1](https://github.com/fabiolb/fabio/tree/v1.7.1) (2026-04-27)

[Full Changelog](https://github.com/fabiolb/fabio/compare/v1.7.0...v1.7.1)

**Implemented enhancements:**

- feat: add ui.path config option for serving UI/API behind a proxy [\#1034](https://github.com/fabiolb/fabio/pull/1034) ([maciej-lech](https://github.com/maciej-lech))

**Fixed bugs:**

- bug: tlsver map missing TLS 1.3 entry causes incorrect Forwarded header for TLS 1.3 connections [\#1029](https://github.com/fabiolb/fabio/issues/1029)

**Closed issues:**

- Missing dependency [\#473](https://github.com/fabiolb/fabio/issues/473)
- Support different base path for UI/API [\#323](https://github.com/fabiolb/fabio/issues/323)

**Merged pull requests:**

- Bump github.com/hashicorp/vault/sdk from 0.25.0 to 0.25.1 [\#1033](https://github.com/fabiolb/fabio/pull/1033) ([dependabot[bot]](https://github.com/apps/dependabot))
- Bump golang.org/x/net from 0.52.0 to 0.53.0 [\#1032](https://github.com/fabiolb/fabio/pull/1032) ([dependabot[bot]](https://github.com/apps/dependabot))
- Bump github.com/hashicorp/consul/api from 1.34.0 to 1.34.2 [\#1031](https://github.com/fabiolb/fabio/pull/1031) ([dependabot[bot]](https://github.com/apps/dependabot))
- fix: add TLS 1.3 to tlsver map and fix comment typo [\#1030](https://github.com/fabiolb/fabio/pull/1030) ([kuishou68](https://github.com/kuishou68))
- Bump github.com/hashicorp/consul/api from 1.33.7 to 1.34.0 [\#1028](https://github.com/fabiolb/fabio/pull/1028) ([dependabot[bot]](https://github.com/apps/dependabot))
- Bump google.golang.org/grpc from 1.79.3 to 1.80.0 [\#1027](https://github.com/fabiolb/fabio/pull/1027) ([dependabot[bot]](https://github.com/apps/dependabot))
- Bump github.com/go-jose/go-jose/v4 from 4.1.3 to 4.1.4 [\#1026](https://github.com/fabiolb/fabio/pull/1026) ([dependabot[bot]](https://github.com/apps/dependabot))
- Bump github.com/hashicorp/vault/api from 1.22.0 to 1.23.0 [\#1024](https://github.com/fabiolb/fabio/pull/1024) ([dependabot[bot]](https://github.com/apps/dependabot))
- Bump github.com/hashicorp/vault/sdk from 0.23.0 to 0.25.0 [\#1022](https://github.com/fabiolb/fabio/pull/1022) ([dependabot[bot]](https://github.com/apps/dependabot))
- Bump github.com/hashicorp/consul/api from 1.33.4 to 1.33.7 [\#1021](https://github.com/fabiolb/fabio/pull/1021) ([dependabot[bot]](https://github.com/apps/dependabot))
- Update Actions and ignore broken win-arm-7 build. [\#1020](https://github.com/fabiolb/fabio/pull/1020) ([tristanmorgan](https://github.com/tristanmorgan))

## [v1.7.0](https://github.com/fabiolb/fabio/tree/v1.7.0) (2026-03-25)

[Full Changelog](https://github.com/fabiolb/fabio/compare/v1.6.11...v1.7.0)

**Implemented enhancements:**

- Remove outdated tracing functions? [\#970](https://github.com/fabiolb/fabio/issues/970)

**Fixed bugs:**

- fix leak for prometheus metrics cleanup [\#1015](https://github.com/fabiolb/fabio/pull/1015) ([evkuzin](https://github.com/evkuzin))

**Closed issues:**

- Metrics are leaking in prometheus [\#979](https://github.com/fabiolb/fabio/issues/979)
- Enhance Fabio to support profiling Tracing  [\#623](https://github.com/fabiolb/fabio/issues/623)

**Merged pull requests:**

- fix: improve proxy addr documentation for multiple protocols use case [\#1018](https://github.com/fabiolb/fabio/pull/1018) ([RodrigoPerestrelo](https://github.com/RodrigoPerestrelo))
- Fix broken links in documentation [\#1017](https://github.com/fabiolb/fabio/pull/1017) ([froque](https://github.com/froque))
- Update GH-Actions and Go version. [\#1016](https://github.com/fabiolb/fabio/pull/1016) ([tristanmorgan](https://github.com/tristanmorgan))
- Remove Tracing and update packages even more. \(\#970\) [\#976](https://github.com/fabiolb/fabio/pull/976) ([tristanmorgan](https://github.com/tristanmorgan))

## [v1.6.11](https://github.com/fabiolb/fabio/tree/v1.6.11) (2025-12-09)

[Full Changelog](https://github.com/fabiolb/fabio/compare/v1.6.10...v1.6.11)

**Merged pull requests:**

- Update dependancies for latest CVEs. [\#1012](https://github.com/fabiolb/fabio/pull/1012) ([tristanmorgan](https://github.com/tristanmorgan))

## [v1.6.10](https://github.com/fabiolb/fabio/tree/v1.6.10) (2025-11-24)

[Full Changelog](https://github.com/fabiolb/fabio/compare/v1.6.9...v1.6.10)

**Closed issues:**

- Call the fatal function within the goroutine of the main test function [\#1009](https://github.com/fabiolb/fabio/issues/1009)
- Support for authentication middleware \(eg: oauth2-proxy\) ? [\#1006](https://github.com/fabiolb/fabio/issues/1006)

**Merged pull requests:**

- Update deps for upstream fixes. [\#1011](https://github.com/fabiolb/fabio/pull/1011) ([tristanmorgan](https://github.com/tristanmorgan))
- Reverse IPv4/IPv6 check. [\#1010](https://github.com/fabiolb/fabio/pull/1010) ([tristanmorgan](https://github.com/tristanmorgan))

## [v1.6.9](https://github.com/fabiolb/fabio/tree/v1.6.9) (2025-10-16)

[Full Changelog](https://github.com/fabiolb/fabio/compare/v1.6.8...v1.6.9)

**Merged pull requests:**

- FROM scratch build and update deps. [\#1008](https://github.com/fabiolb/fabio/pull/1008) ([tristanmorgan](https://github.com/tristanmorgan))
- feat: add armv7 support for docker images [\#1007](https://github.com/fabiolb/fabio/pull/1007) ([amd989](https://github.com/amd989))

## [v1.6.8](https://github.com/fabiolb/fabio/tree/v1.6.8) (2025-09-23)

[Full Changelog](https://github.com/fabiolb/fabio/compare/v1.6.7...v1.6.8)

**Closed issues:**

- Health check on port 9999 [\#995](https://github.com/fabiolb/fabio/issues/995)
- Multiple entries in proxy.auth do not work as specified in documentation [\#929](https://github.com/fabiolb/fabio/issues/929)
- Correct Consul ACL Policy for Fabio [\#831](https://github.com/fabiolb/fabio/issues/831)
- Translating `upstream-host/path` to `path.example.com` [\#801](https://github.com/fabiolb/fabio/issues/801)

**Merged pull requests:**

- Package dependencies updated. [\#1004](https://github.com/fabiolb/fabio/pull/1004) ([tristanmorgan](https://github.com/tristanmorgan))
- docs: Fixes bad example for creating multiple basic authorization schemes [\#1003](https://github.com/fabiolb/fabio/pull/1003) ([steffkelsey](https://github.com/steffkelsey))
- Enabling many more linters in the pipeline. [\#999](https://github.com/fabiolb/fabio/pull/999) ([tristanmorgan](https://github.com/tristanmorgan))
- Fix missing brand logo in routes page [\#998](https://github.com/fabiolb/fabio/pull/998) ([tristanmorgan](https://github.com/tristanmorgan))
- Fixing up some web links [\#997](https://github.com/fabiolb/fabio/pull/997) ([tristanmorgan](https://github.com/tristanmorgan))
- Fix the ui.routingtable.source.newtab doc. [\#996](https://github.com/fabiolb/fabio/pull/996) ([tristanmorgan](https://github.com/tristanmorgan))
- extract only binary from zipfile [\#994](https://github.com/fabiolb/fabio/pull/994) ([shantanugadgil](https://github.com/shantanugadgil))
- add option to constrain fabio instance to specific consul namespace [\#812](https://github.com/fabiolb/fabio/pull/812) ([baabgai](https://github.com/baabgai))

## [v1.6.7](https://github.com/fabiolb/fabio/tree/v1.6.7) (2025-05-30)

[Full Changelog](https://github.com/fabiolb/fabio/compare/v1.6.6...v1.6.7)

**Merged pull requests:**

- Bump go.mod pins to latest point release. [\#993](https://github.com/fabiolb/fabio/pull/993) ([tristanmorgan](https://github.com/tristanmorgan))

## [v1.6.6](https://github.com/fabiolb/fabio/tree/v1.6.6) (2025-05-26)

[Full Changelog](https://github.com/fabiolb/fabio/compare/v1.6.5...v1.6.6)

**Implemented enhancements:**

- Add a golang-linter to CI [\#972](https://github.com/fabiolb/fabio/pull/972) ([aleksraiden](https://github.com/aleksraiden))

**Closed issues:**

- wiki content vs fabio/docs/content ? [\#986](https://github.com/fabiolb/fabio/issues/986)

**Merged pull requests:**

- Using staticcheck to fix many issues. [\#992](https://github.com/fabiolb/fabio/pull/992) ([tristanmorgan](https://github.com/tristanmorgan))
- Update golangci-lint and run yamlfmt. [\#990](https://github.com/fabiolb/fabio/pull/990) ([tristanmorgan](https://github.com/tristanmorgan))
- Fix mistake made in \#988. [\#989](https://github.com/fabiolb/fabio/pull/989) ([tristanmorgan](https://github.com/tristanmorgan))
- Actions permissions [\#988](https://github.com/fabiolb/fabio/pull/988) ([tristanmorgan](https://github.com/tristanmorgan))
- docs: fix broken link [\#987](https://github.com/fabiolb/fabio/pull/987) ([marco-m](https://github.com/marco-m))
- Update dependancies including GoBGP. [\#985](https://github.com/fabiolb/fabio/pull/985) ([tristanmorgan](https://github.com/tristanmorgan))
- Add a CODEOWNERS file. [\#983](https://github.com/fabiolb/fabio/pull/983) ([tristanmorgan](https://github.com/tristanmorgan))
- Document insensitive prefix matching in the list. [\#982](https://github.com/fabiolb/fabio/pull/982) ([tristanmorgan](https://github.com/tristanmorgan))
- Update golang.org/x/net. [\#980](https://github.com/fabiolb/fabio/pull/980) ([tristanmorgan](https://github.com/tristanmorgan))

## [v1.6.5](https://github.com/fabiolb/fabio/tree/v1.6.5) (2025-02-28)

[Full Changelog](https://github.com/fabiolb/fabio/compare/v1.6.4...v1.6.5)

**Implemented enhancements:**

- Unable to load correct certificates if 1 invalid one is in consul k/v [\#941](https://github.com/fabiolb/fabio/issues/941)
- Use a Go 1.24.0 [\#971](https://github.com/fabiolb/fabio/pull/971) ([aleksraiden](https://github.com/aleksraiden))
- Report all certificate errors instead of stopping at the first. \(\#941\) [\#964](https://github.com/fabiolb/fabio/pull/964) ([tristanmorgan](https://github.com/tristanmorgan))

**Closed issues:**

- Please bump golang.org/x/sys dependency to enable a build on riscv64-freebsd [\#927](https://github.com/fabiolb/fabio/issues/927)
- Fabio is using Datadog reserved tag keys  [\#923](https://github.com/fabiolb/fabio/issues/923)

**Merged pull requests:**

- Update deps to latest [\#975](https://github.com/fabiolb/fabio/pull/975) ([aleksraiden](https://github.com/aleksraiden))
- updating godeps [\#969](https://github.com/fabiolb/fabio/pull/969) ([aleksraiden](https://github.com/aleksraiden))
- Update Hugo config to work with version bump in \#965 [\#967](https://github.com/fabiolb/fabio/pull/967) ([tristanmorgan](https://github.com/tristanmorgan))
- Use Alpine3.21 as base docker image [\#966](https://github.com/fabiolb/fabio/pull/966) ([aleksraiden](https://github.com/aleksraiden))
- update CI components [\#965](https://github.com/fabiolb/fabio/pull/965) ([aleksraiden](https://github.com/aleksraiden))
- README: remove mention to www.kijiji.it \(decommissioned in 2022\) [\#963](https://github.com/fabiolb/fabio/pull/963) ([marco-m-pix4d](https://github.com/marco-m-pix4d))
- Update dependancies again. [\#962](https://github.com/fabiolb/fabio/pull/962) ([tristanmorgan](https://github.com/tristanmorgan))
- Use ParseUint to test for overflow directly [\#961](https://github.com/fabiolb/fabio/pull/961) ([dcarbone](https://github.com/dcarbone))
- Fix small typo in DogStatsD config ref. [\#960](https://github.com/fabiolb/fabio/pull/960) ([tristanmorgan](https://github.com/tristanmorgan))
- Rebuild CHANGELOG [\#959](https://github.com/fabiolb/fabio/pull/959) ([tristanmorgan](https://github.com/tristanmorgan))
- Adds handling of Datadog reserved tag keys [\#924](https://github.com/fabiolb/fabio/pull/924) ([froque](https://github.com/froque))

## [v1.6.4](https://github.com/fabiolb/fabio/tree/v1.6.4) (2024-11-27)

[Full Changelog](https://github.com/fabiolb/fabio/compare/v1.6.3...v1.6.4)

**Closed issues:**

- CI pipeline to run testsuite [\#949](https://github.com/fabiolb/fabio/issues/949)
- Fabio not exporting all the metrics with prometheus [\#947](https://github.com/fabiolb/fabio/issues/947)
- certificates - cert and ca chain/intermediate [\#946](https://github.com/fabiolb/fabio/issues/946)
- This repository is unmaintaned. [\#944](https://github.com/fabiolb/fabio/issues/944)
- CVE-2023-44487 HTTP/2 rapid reset [\#939](https://github.com/fabiolb/fabio/issues/939)
- TCP no route - cant balance tcp [\#936](https://github.com/fabiolb/fabio/issues/936)
- windows: setting logging path in fabio properties [\#920](https://github.com/fabiolb/fabio/issues/920)
- Port range in the proxy.addr [\#529](https://github.com/fabiolb/fabio/issues/529)

**Merged pull requests:**

- Add GoReleaser workflow. [\#958](https://github.com/fabiolb/fabio/pull/958) ([tristanmorgan](https://github.com/tristanmorgan))
- go-kit/kit/log go-kit/log [\#956](https://github.com/fabiolb/fabio/pull/956) ([tristanmorgan](https://github.com/tristanmorgan))
- Update go-retryablehttp to fix warning. [\#954](https://github.com/fabiolb/fabio/pull/954) ([tristanmorgan](https://github.com/tristanmorgan))
- Update and try fix GH Pages publish action. [\#953](https://github.com/fabiolb/fabio/pull/953) ([tristanmorgan](https://github.com/tristanmorgan))
- Update go version, test binaries and package versions. [\#952](https://github.com/fabiolb/fabio/pull/952) ([tristanmorgan](https://github.com/tristanmorgan))
- Remove vendored modules in favour of go mod. [\#951](https://github.com/fabiolb/fabio/pull/951) ([tristanmorgan](https://github.com/tristanmorgan))
- Update Github runner image. [\#950](https://github.com/fabiolb/fabio/pull/950) ([tristanmorgan](https://github.com/tristanmorgan))
- fix: close resp body [\#945](https://github.com/fabiolb/fabio/pull/945) ([testwill](https://github.com/testwill))
- Trim leading and trailing spaces from service tags [\#943](https://github.com/fabiolb/fabio/pull/943) ([logocomune](https://github.com/logocomune))
- Fix doubled download a Vault file [\#942](https://github.com/fabiolb/fabio/pull/942) ([aleksraiden](https://github.com/aleksraiden))
- Remove deprecated ioutil [\#940](https://github.com/fabiolb/fabio/pull/940) ([tristanmorgan](https://github.com/tristanmorgan))
- Dockerfile: add CAP\_NET\_BIND\_SERVICE+eip to fabio to allow running as root [\#938](https://github.com/fabiolb/fabio/pull/938) ([Kamilcuk](https://github.com/Kamilcuk))
- Consul registry performance improvements [\#928](https://github.com/fabiolb/fabio/pull/928) ([ddreier](https://github.com/ddreier))
- \[Docs\] Fix wrong parameter name [\#914](https://github.com/fabiolb/fabio/pull/914) ([KEANO89](https://github.com/KEANO89))
- Updating grpc handler to gracefully close backend connections [\#913](https://github.com/fabiolb/fabio/pull/913) ([nathanejohnson](https://github.com/nathanejohnson))

## [v1.6.3](https://github.com/fabiolb/fabio/tree/v1.6.3) (2022-12-09)

[Full Changelog](https://github.com/fabiolb/fabio/compare/v1.6.2...v1.6.3)

**Implemented enhancements:**

- Feature request: Make source links in ui interface clickable [\#901](https://github.com/fabiolb/fabio/issues/901)

**Closed issues:**

- Ignore host=dst when backend is https [\#916](https://github.com/fabiolb/fabio/issues/916)
- poll new feature requests [\#910](https://github.com/fabiolb/fabio/issues/910)
- Fabio Clustering. [\#668](https://github.com/fabiolb/fabio/issues/668)

**Merged pull requests:**

- Disable BGP functionality on Windows since gobgp does not support this. [\#919](https://github.com/fabiolb/fabio/pull/919) ([nathanejohnson](https://github.com/nathanejohnson))
- updating CHANGELOG [\#918](https://github.com/fabiolb/fabio/pull/918) ([nathanejohnson](https://github.com/nathanejohnson))
- Don't use "dst" literal as sni name when host=dst is specified on https backends [\#917](https://github.com/fabiolb/fabio/pull/917) ([nathanejohnson](https://github.com/nathanejohnson))
- add feature to advertise anycast addresses via BGP [\#909](https://github.com/fabiolb/fabio/pull/909) ([nathanejohnson](https://github.com/nathanejohnson))
- Change the shutdown procedure to deregister fabio from the registry and then shutdown the proxy [\#908](https://github.com/fabiolb/fabio/pull/908) ([martinivanov](https://github.com/martinivanov))
- Feature/source link [\#907](https://github.com/fabiolb/fabio/pull/907) ([KTruesdellENA](https://github.com/KTruesdellENA))

## [v1.6.2](https://github.com/fabiolb/fabio/tree/v1.6.2) (2022-09-13)

[Full Changelog](https://github.com/fabiolb/fabio/compare/v1.6.1...v1.6.2)

**Closed issues:**

- Update TLS cipher parser to include modern ciphers [\#903](https://github.com/fabiolb/fabio/issues/903)
- Custom behavior for the situation when the service has no healthy instances [\#898](https://github.com/fabiolb/fabio/issues/898)

**Merged pull requests:**

- update README for v1.6.2 release [\#905](https://github.com/fabiolb/fabio/pull/905) ([nathanejohnson](https://github.com/nathanejohnson))
- Updating TLS cipher config parser to include TLS 1.3 constants. [\#904](https://github.com/fabiolb/fabio/pull/904) ([nathanejohnson](https://github.com/nathanejohnson))

## [v1.6.1](https://github.com/fabiolb/fabio/tree/v1.6.1) (2022-07-19)

[Full Changelog](https://github.com/fabiolb/fabio/compare/v1.6.0...v1.6.1)

**Implemented enhancements:**

- Multi-DC fabio [\#115](https://github.com/fabiolb/fabio/issues/115)

**Fixed bugs:**

- Crash: invalid log msg: http2: panic serving CLIENT\_IP:CLIENT\_PORT: runtime error: index out of range \[-1\] [\#872](https://github.com/fabiolb/fabio/issues/872)

**Closed issues:**

- admin UI Overrides not working [\#886](https://github.com/fabiolb/fabio/issues/886)
- Panic on created prometheus metric name [\#878](https://github.com/fabiolb/fabio/issues/878)
- Crash on route update: panic: runtime error: index out of range, diffmatchpatch.\(\*DiffMatchPatch\).DiffCharsToLines [\#873](https://github.com/fabiolb/fabio/issues/873)
- Experiencing 502's [\#862](https://github.com/fabiolb/fabio/issues/862)
- Fabio immediately drop routes when consul agent unavailable for a while [\#861](https://github.com/fabiolb/fabio/issues/861)
- \[proxy/tls\] Update supported TLS versions and cipher suites [\#858](https://github.com/fabiolb/fabio/issues/858)
- JSON schema is incorrect in website Dest should be Dst [\#852](https://github.com/fabiolb/fabio/issues/852)
- \[question\] URL for TLS destination [\#850](https://github.com/fabiolb/fabio/issues/850)
- \[Feature\] Possibility of adding arm/arm64 docker builds. [\#833](https://github.com/fabiolb/fabio/issues/833)

**Merged pull requests:**

- Release/v1.6.1 [\#897](https://github.com/fabiolb/fabio/pull/897) ([nathanejohnson](https://github.com/nathanejohnson))
- setting sni to match host [\#896](https://github.com/fabiolb/fabio/pull/896) ([KTruesdellENA](https://github.com/KTruesdellENA))
- Update random picker to use math/rand's Intn function [\#893](https://github.com/fabiolb/fabio/pull/893) ([nathanejohnson](https://github.com/nathanejohnson))
- add configurable grpc message sizes to \#632 [\#890](https://github.com/fabiolb/fabio/pull/890) ([nathanejohnson](https://github.com/nathanejohnson))
- add tls13 [\#889](https://github.com/fabiolb/fabio/pull/889) ([nathanejohnson](https://github.com/nathanejohnson))
- update materialize bits. see issue \#886 [\#888](https://github.com/fabiolb/fabio/pull/888) ([nathanejohnson](https://github.com/nathanejohnson))
- Moved admin UI assets to use go embed [\#885](https://github.com/fabiolb/fabio/pull/885) ([nathanejohnson](https://github.com/nathanejohnson))
- update the custom css [\#884](https://github.com/fabiolb/fabio/pull/884) ([KTruesdellENA](https://github.com/KTruesdellENA))
- Bump go-diff dependency version to 1.2.0.  Fixes \#873 [\#881](https://github.com/fabiolb/fabio/pull/881) ([nathanejohnson](https://github.com/nathanejohnson))
- bump HUGO version to 0.101.0 [\#880](https://github.com/fabiolb/fabio/pull/880) ([nathanejohnson](https://github.com/nathanejohnson))
- add docs from PR \#854 to fabio.properties [\#879](https://github.com/fabiolb/fabio/pull/879) ([nathanejohnson](https://github.com/nathanejohnson))
- Build multi-arch Docker images for amd64 and arm64 architectures [\#875](https://github.com/fabiolb/fabio/pull/875) ([vamc19](https://github.com/vamc19))
- Fix x-forwarded-for header processing for ws connections [\#860](https://github.com/fabiolb/fabio/pull/860) ([bn0ir](https://github.com/bn0ir))
- Update registry.backend.md [\#854](https://github.com/fabiolb/fabio/pull/854) ([webmutation](https://github.com/webmutation))
- Resulting complete routing table has no 'tags "a,b"'  in last line [\#841](https://github.com/fabiolb/fabio/pull/841) ([hb9cwp](https://github.com/hb9cwp))
- fixes \#807 - changes map for grpc connections to be a string key [\#816](https://github.com/fabiolb/fabio/pull/816) ([nathanejohnson](https://github.com/nathanejohnson))
- Add command line flag to toggle required consistency on consul reads [\#811](https://github.com/fabiolb/fabio/pull/811) ([jeremycw](https://github.com/jeremycw))
- Issue 605 grpc host matching [\#632](https://github.com/fabiolb/fabio/pull/632) ([tommyalatalo](https://github.com/tommyalatalo))

## [v1.6.0](https://github.com/fabiolb/fabio/tree/v1.6.0) (2022-04-11)

[Full Changelog](https://github.com/fabiolb/fabio/compare/v1.5.15...v1.6.0)

**Implemented enhancements:**

- Add support for influxdb metrics [\#253](https://github.com/fabiolb/fabio/issues/253)
- Support for prometheus [\#211](https://github.com/fabiolb/fabio/issues/211)
- Support dogstatd with tags [\#165](https://github.com/fabiolb/fabio/issues/165)
- Riemann metrics support [\#126](https://github.com/fabiolb/fabio/issues/126)
- Simple HTTP path prefix replacement [\#767](https://github.com/fabiolb/fabio/pull/767) ([JamesJJ](https://github.com/JamesJJ))

**Closed issues:**

- Consul Route updates very slow with large numbers of routes [\#865](https://github.com/fabiolb/fabio/issues/865)
- Restricting TLS versions [\#859](https://github.com/fabiolb/fabio/issues/859)
- \[admin/ui\] General updates [\#856](https://github.com/fabiolb/fabio/issues/856)
- \[question\] - Can Fabio listen on 80/tcp with Nomad [\#844](https://github.com/fabiolb/fabio/issues/844)
- Supporting requests in the form of /my-app/page1 [\#842](https://github.com/fabiolb/fabio/issues/842)
- Fabio Using Container IPs to create routes [\#839](https://github.com/fabiolb/fabio/issues/839)
- All my dynamic routes suddenly vanished! [\#837](https://github.com/fabiolb/fabio/issues/837)
- Fabio redirecting to /routes on own service [\#832](https://github.com/fabiolb/fabio/issues/832)
- docu fabio configure TLS/SSL\(HTTPS\) understanding problem [\#827](https://github.com/fabiolb/fabio/issues/827)
- Crash: \[FATAL\] 1.5.13. strconv.ParseUint: parsing ":1883;PROTO=TCP-DYNAMIC": invalid syntax [\#826](https://github.com/fabiolb/fabio/issues/826)
- strip doesn't work as expected on redirect [\#824](https://github.com/fabiolb/fabio/issues/824)
- Using Fabio with Consul over mTLS [\#820](https://github.com/fabiolb/fabio/issues/820)
- Switch to github actions [\#817](https://github.com/fabiolb/fabio/issues/817)
- Panic - httputil: ReverseProxy read error during body copy [\#814](https://github.com/fabiolb/fabio/issues/814)
- Support for Consul and Vault Namespaces [\#810](https://github.com/fabiolb/fabio/issues/810)
- grpc be closed when uninstall service target [\#807](https://github.com/fabiolb/fabio/issues/807)
- fabio binary filename for download [\#805](https://github.com/fabiolb/fabio/issues/805)
- Add arm64 arch [\#804](https://github.com/fabiolb/fabio/issues/804)
- Can Fabio to prefer one ethernet interface for proxy\_addr? [\#802](https://github.com/fabiolb/fabio/issues/802)
- TCP Dynamic Proxy route without specifying exact IP? [\#797](https://github.com/fabiolb/fabio/issues/797)
- \[Question\] What are opinions on allowing stale reads of Consul Catalog [\#764](https://github.com/fabiolb/fabio/issues/764)
- Simple HTTP path prefix replacement [\#691](https://github.com/fabiolb/fabio/issues/691)
- Does Fabio support multiple CS Stores per listener? [\#666](https://github.com/fabiolb/fabio/issues/666)
- \[Question\] Stats - Status code per service [\#371](https://github.com/fabiolb/fabio/issues/371)
- Statsd output is not good [\#327](https://github.com/fabiolb/fabio/issues/327)
- Send metrics to cloudwatch [\#326](https://github.com/fabiolb/fabio/issues/326)
- Mixing of HTTPS proxying and SNI+TCP on a single port [\#213](https://github.com/fabiolb/fabio/issues/213)

**Merged pull requests:**

- gofmt [\#870](https://github.com/fabiolb/fabio/pull/870) ([nathanejohnson](https://github.com/nathanejohnson))
- updating x/sys [\#869](https://github.com/fabiolb/fabio/pull/869) ([nathanejohnson](https://github.com/nathanejohnson))
- update go and alpine versions [\#868](https://github.com/fabiolb/fabio/pull/868) ([Netlims](https://github.com/Netlims))
- \#865 Move the route table sort into NewTable so that it only happens once. [\#867](https://github.com/fabiolb/fabio/pull/867) ([ddreier](https://github.com/ddreier))
- removing exclusion of arm64 mac build.  Fixes \#804 [\#866](https://github.com/fabiolb/fabio/pull/866) ([nathanejohnson](https://github.com/nathanejohnson))
- Fix example commands in registry.consul.kvpath [\#864](https://github.com/fabiolb/fabio/pull/864) ([blake](https://github.com/blake))
- Add IdleConnTimeout configurable for http transport [\#863](https://github.com/fabiolb/fabio/pull/863) ([aal89](https://github.com/aal89))
- admin/ui updates: [\#857](https://github.com/fabiolb/fabio/pull/857) ([dcarbone](https://github.com/dcarbone))
- Update 2 broken links in documentation [\#822](https://github.com/fabiolb/fabio/pull/822) ([mig4ng](https://github.com/mig4ng))
- Fix small typo in README.md [\#821](https://github.com/fabiolb/fabio/pull/821) ([mig4ng](https://github.com/mig4ng))
- Add support for github actions [\#819](https://github.com/fabiolb/fabio/pull/819) ([nathanejohnson](https://github.com/nathanejohnson))
- Remove golang toolchain name from release binary names [\#818](https://github.com/fabiolb/fabio/pull/818) ([nathanejohnson](https://github.com/nathanejohnson))
- we don't use Fabio [\#813](https://github.com/fabiolb/fabio/pull/813) ([hsmade](https://github.com/hsmade))
- Updating tcp dynamic proxy to match on routes that are port only [\#806](https://github.com/fabiolb/fabio/pull/806) ([nathanejohnson](https://github.com/nathanejohnson))
- Refactor metrics [\#476](https://github.com/fabiolb/fabio/pull/476) ([magiconair](https://github.com/magiconair))



\* *This Changelog was automatically generated by [github_changelog_generator](https://github.com/github-changelog-generator/github-changelog-generator)*

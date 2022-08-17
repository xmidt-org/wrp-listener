# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.2.5]
- Patched Vulnerabilities
  - https://github.com/xmidt-org/wrp-listener/issues/50
  - https://github.com/xmidt-org/wrp-listener/issues/49
- Dependency update
  - Introduces new vuln https://www.mend.io/vulnerability-database/CVE-2022-29526
  - Introduces new vuln https://www.mend.io/vulnerability-database/CVE-2021-38561

## [v0.2.4]
- bump dependencies [#48](https://github.com/xmidt-org/wrp-listener/pull/48)

## [v0.2.3]
- Fixed hashTokenFactory build error. [#40](https://github.com/xmidt-org/wrp-listener/pull/40)

## [v0.2.2]
- Add interval check and create shutdown channel in `NewPeriodicRegisterer` [#30](https://github.com/xmidt-org/wrp-listener/pull/30)
- Add provide functions for webhook client metrics and periodic registerer [#38](https://github.com/xmidt-org/wrp-listener/pull/38)

## [v0.2.1]
- Downgraded uuid version [#19](https://github.com/xmidt-org/wrp-listener/pull/19)

## [v0.2.0]
- Added metrics for periodicRegisterer [#18](https://github.com/xmidt-org/wrp-listener/pull/18)
- Added parser for extracting device ID from wrp message [#23](https://github.com/xmidt-org/wrp-listener/pull/23)

## [v0.1.2]
- Added travis automation for github releases [#13](https://github.com/xmidt-org/wrp-listener/pull/13)
- bumped bascule version [#16](https://github.com/xmidt-org/wrp-listener/pull/16)

## [v0.1.1]
- Fixed go-kit version

## [v0.1.0]
- Initial creation, moved from: https://github.com/xmidt-org/svalinn
- Modified authentication to work with `bascule` package/repo
- Refactored registerers

[Unreleased]: https://github.com/xmidt-org/wrp-listener/compare/v0.2.5..HEAD
[v0.2.5]: https://github.com/xmidt-org/wrp-listener/compare/v0.2.4..v0.2.5
[v0.2.4]: https://github.com/xmidt-org/wrp-listener/compare/v0.2.3..v0.2.4
[v0.2.3]: https://github.com/xmidt-org/wrp-listener/compare/v0.2.2..v0.2.3
[v0.2.2]: https://github.com/xmidt-org/wrp-listener/compare/v0.2.1..v0.2.2
[v0.2.1]: https://github.com/xmidt-org/wrp-listener/compare/v0.2.0..v0.2.1
[v0.2.0]: https://github.com/xmidt-org/wrp-listener/compare/v0.1.2..v0.2.0
[v0.1.2]: https://github.com/xmidt-org/wrp-listener/compare/v0.1.1..v0.1.2
[v0.1.1]: https://github.com/xmidt-org/wrp-listener/compare/0.1.0...v0.1.1
[v0.1.0]: https://github.com/xmidt-org/wrp-listener/compare/0.0.0...v0.1.0

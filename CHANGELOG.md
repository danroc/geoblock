# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2025-12-06

### Changed

- Change logging library and format ([#272](https://github.com/danroc/geoblock/pull/272))

### Removed

- **BREAKING:** Remove `/v1/metrics` endpoint (JSON format). Use `/metrics` (Prometheus format) ([#279](https://github.com/danroc/geoblock/pull/279), [#280](https://github.com/danroc/geoblock/pull/280))

## [0.3.3] - 2025-10-05

### Added

- Add Prometheus metrics support ([#237](https://github.com/danroc/geoblock/pull/237))

## [0.3.2] - 2025-08-24

### Changed

- Combine `request_allowed` and `request_valid` log fields into `request_status` ([#223](https://github.com/danroc/geoblock/pull/223))

## [0.3.1] - 2025-08-20

### Added

- Log whether requests are valid and allowed ([#199](https://github.com/danroc/geoblock/pull/199))

## [0.3.0] - 2025-08-20

### Changed

- **BREAKING:** Log in JSON format by default ([#195](https://github.com/danroc/geoblock/pull/195))

## [0.2.1] - 2025-08-14

### Changed

- **Note:** This release contains only internal refactoring and dependency updates.

## [0.2.0] - 2025-07-30

### Changed

- **BREAKING:** Change the object returned by the `/metrics` endpoint ([#181](https://github.com/danroc/geoblock/pull/181))

## [0.1.23] - 2025-07-28

### Added

- Log version during initialization ([#164](https://github.com/danroc/geoblock/pull/164))

## [0.1.22] - 2025-07-28

### Fixed

- Set a timeout when fetching updates for the IP database ([#162](https://github.com/danroc/geoblock/pull/162))

## [0.1.21] - 2025-07-25

### Added

- Log whether the source IP is from a local network ([#155](https://github.com/danroc/geoblock/pull/155))

## [0.1.20] - 2025-07-11

### Changed

- Improve error logging ([#148](https://github.com/danroc/geoblock/pull/148))

### Fixed

- Prevent possible race condition when computing the `total` metric ([#136](https://github.com/danroc/geoblock/pull/136))

## [0.1.19] - 2025-05-31

### Changed

- **Note:** This release contains only internal refactoring and dependency updates.

## [0.1.18] - 2025-04-08

### Changed

- **Note:** This release contains only internal refactoring and dependency updates.

## [0.1.17] - 2025-02-24

### Changed

- **Note:** This release contains only internal refactoring and dependency updates.

## [0.1.16] - 2025-01-09

### Added

- Add a license ([#76](https://github.com/danroc/geoblock/pull/76))

### Removed

- Don't log redundant error message ([#71](https://github.com/danroc/geoblock/pull/71))

## [0.1.15] - 2024-12-07

### Changed

- **Note:** This release contains only internal refactoring and dependency updates.

## [0.1.14] - 2024-11-28

### Changed

- **Note:** This release contains only internal refactoring and dependency updates.

## [0.1.13] - 2024-11-28

### Changed

- Optimize IP database (use interval tree) ([#47](https://github.com/danroc/geoblock/pull/47))

## [0.1.12] - 2024-11-23

### Added

- Allow wildcard domains in config ([#32](https://github.com/danroc/geoblock/pull/32), [#42](https://github.com/danroc/geoblock/pull/42))

### Changed

- Rename `requested_*` to `request_*` in logs ([#39](https://github.com/danroc/geoblock/pull/39))

## [0.1.11]

### Changed

- Increase healthcheck timeouts ([#25](https://github.com/danroc/geoblock/pull/25))
- Change default configuration path ([#29](https://github.com/danroc/geoblock/pull/29))

## [0.1.10] - 2024-11-17

### Added

- Add `/metrics` endpoint ([#19](https://github.com/danroc/geoblock/pull/19))

## [0.1.9] - 2024-11-16

### Added

- Support setting log level ([#9](https://github.com/danroc/geoblock/pull/9))

### Changed

- Run container as a non-root user ([#11](https://github.com/danroc/geoblock/pull/11), [#15](https://github.com/danroc/geoblock/pull/15))

## [0.1.8] - 2024-11-14

### Added

- Auto-update databases and auto-reload configuration ([#8](https://github.com/danroc/geoblock/pull/8))

## [0.1.7] - 2024-11-12

### Added

- Support method filtering ([#3](https://github.com/danroc/geoblock/pull/3))

### Fixed

- Handle the case where the source IP is invalid
- Log errors at the right level

## [0.1.6] - 2024-10-31

### Added

- Add health-check to Docker image

## [0.1.5] - 2024-10-31

### Changed

- Revert "Add health-check to Docker image"

## [0.1.4] - 2024-10-31

### Added

- Add health-check to Docker image
- Add health-check endpoint

## [0.1.3] - 2024-10-31

### Changed

- **Note:** This release contains only internal refactoring and dependency updates.

## [0.1.2] - 2024-10-26

### Changed

- Change databases to use GeoLite2 only

## [0.1.1] - 2024-10-26

### Changed

- **Note:** This release contains only internal refactoring and dependency updates.

## [0.1.0] - 2024-10-25

### Added

- Add timeouts to HTTP server
- Add rules engine
- Add CIDR unmarshalling and validation
- Add autonomous systems to configuration

[Unreleased]: https://github.com/danroc/geoblock/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/danroc/geoblock/compare/v0.3.3...v0.4.0
[0.3.3]: https://github.com/danroc/geoblock/compare/v0.3.2...v0.3.3
[0.3.2]: https://github.com/danroc/geoblock/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/danroc/geoblock/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/danroc/geoblock/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/danroc/geoblock/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/danroc/geoblock/compare/v0.1.23...v0.2.0
[0.1.23]: https://github.com/danroc/geoblock/compare/v0.1.22...v0.1.23
[0.1.22]: https://github.com/danroc/geoblock/compare/v0.1.21...v0.1.22
[0.1.21]: https://github.com/danroc/geoblock/compare/v0.1.20...v0.1.21
[0.1.20]: https://github.com/danroc/geoblock/compare/v0.1.19...v0.1.20
[0.1.19]: https://github.com/danroc/geoblock/compare/v0.1.18...v0.1.19
[0.1.18]: https://github.com/danroc/geoblock/compare/v0.1.17...v0.1.18
[0.1.17]: https://github.com/danroc/geoblock/compare/v0.1.16...v0.1.17
[0.1.16]: https://github.com/danroc/geoblock/compare/v0.1.15...v0.1.16
[0.1.15]: https://github.com/danroc/geoblock/compare/v0.1.14...v0.1.15
[0.1.14]: https://github.com/danroc/geoblock/compare/v0.1.13...v0.1.14
[0.1.13]: https://github.com/danroc/geoblock/compare/v0.1.12...v0.1.13
[0.1.12]: https://github.com/danroc/geoblock/compare/v0.1.11...v0.1.12
[0.1.11]: https://github.com/danroc/geoblock/compare/v0.1.10...v0.1.11
[0.1.10]: https://github.com/danroc/geoblock/compare/v0.1.9...v0.1.10
[0.1.9]: https://github.com/danroc/geoblock/compare/v0.1.8...v0.1.9
[0.1.8]: https://github.com/danroc/geoblock/compare/v0.1.7...v0.1.8
[0.1.7]: https://github.com/danroc/geoblock/compare/v0.1.6...v0.1.7
[0.1.6]: https://github.com/danroc/geoblock/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/danroc/geoblock/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/danroc/geoblock/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/danroc/geoblock/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/danroc/geoblock/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/danroc/geoblock/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/danroc/geoblock/releases/tag/v0.1.0

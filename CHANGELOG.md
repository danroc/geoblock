# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.15] - 2024-12-07

### Changed

- CI improvements

## [0.1.14] - 2024-11-28

### Changed

- CI improvements

## [0.1.13] - 2024-11-28

### Changed

- Use interval tree to store database entries ([#47](https://github.com/danroc/geoblock/pull/47))

## [0.1.12] - 2024-11-23

### Added

- Allow wildcard domains in config ([#32](https://github.com/danroc/geoblock/pull/32), [#42](https://github.com/danroc/geoblock/pull/42))

### Changed

- Rename `requested_*` to `request_*` in logs ([#39](https://github.com/danroc/geoblock/pull/39))

## [0.1.11]

### Changed

- Increase `HEALTHCHECK` timeouts ([#25](https://github.com/danroc/geoblock/pull/25))
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

## [0.1.2] - 2024-10-26

### Changed

- Change databases to use GeoLite2 only

## [0.1.1] - 2024-10-26

## [0.1.0] - 2024-10-25

### Added

- Add timeouts to HTTP server
- Add rules engine
- Add CIDR unmarshalling and validation
- Add autonomous systems to configuration

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

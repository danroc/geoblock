# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Run container as a non-root user ([#11](https://github.com/danroc/geoblock/pull/11))
- Support setting log level ([#9](https://github.com/danroc/geoblock/pull/9))

## [v0.1.8] - 2024-11-14

### Added

- Auto-update databases and auto-reload configuration ([#8](https://github.com/danroc/geoblock/pull/8))

## [v0.1.7] - 2024-11-12

### Added

- Support method filtering ([#3](https://github.com/danroc/geoblock/pull/3))

### Fixed

- Handle the case where the source IP is invalid
- Log errors at the right level

## [v0.1.6] - 2024-10-31

### Added

- Add health-check to Docker image

## [v0.1.5] - 2024-10-31

### Changed

- Revert "Add health-check to Docker image"

## [v0.1.4] - 2024-10-31

### Added

- Add health-check to Docker image
- Add health-check endpoint

## [v0.1.3] - 2024-10-31

## [v0.1.2] - 2024-10-26

### Changed

- Change databases to use GeoLite2 only

## [v0.1.1] - 2024-10-26

## [v0.1.0] - 2024-10-25

### Added

- Add timeouts to HTTP server
- Add rules engine
- Add CIDR unmarshalling and validation
- Add autonomous systems to configuration
- Add duration parsing

[Unreleased]: https://github.com/danroc/geoblock/compare/v0.1.8...HEAD
[v0.1.8]: https://github.com/danroc/geoblock/compare/v0.1.7...v0.1.8
[v0.1.7]: https://github.com/danroc/geoblock/compare/v0.1.6...v0.1.7
[v0.1.6]: https://github.com/danroc/geoblock/compare/v0.1.5...v0.1.6
[v0.1.5]: https://github.com/danroc/geoblock/compare/v0.1.4...v0.1.5
[v0.1.4]: https://github.com/danroc/geoblock/compare/v0.1.3...v0.1.4
[v0.1.3]: https://github.com/danroc/geoblock/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/danroc/geoblock/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/danroc/geoblock/compare/v0.1.0...v0.1.1
[v0.1.0]: https://github.com/danroc/geoblock/releases/tag/v0.1.0

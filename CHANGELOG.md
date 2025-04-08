# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.18] - 2025-04-08

### Changed

- Update module golang.org/x/net to v0.39.0 ([#122](https://github.com/danroc/geoblock/pull/122))
- Update module golang.org/x/crypto to v0.37.0 ([#121](https://github.com/danroc/geoblock/pull/121))
- Update golang docker tag to v1.24.2 ([#118](https://github.com/danroc/geoblock/pull/118))
- Update module github.com/go-playground/validator/v10 to v10.26.0 ([#117](https://github.com/danroc/geoblock/pull/117))
- Update module golang.org/x/net to v0.38.0 ([#116](https://github.com/danroc/geoblock/pull/116))
- Update module golang.org/x/net to v0.37.0 ([#115](https://github.com/danroc/geoblock/pull/115))
- Update module golang.org/x/net to v0.36.0 ([#111](https://github.com/danroc/geoblock/pull/111))
- Update golang docker tag to v1.24.1 ([#110](https://github.com/danroc/geoblock/pull/110))
- Update module golang.org/x/crypto to v0.35.0 ([#108](https://github.com/danroc/geoblock/pull/108))

## [0.1.17] - 2025-02-24

### Changed

- Update module golang.org/x/crypto to v0.34.0 ([#107](https://github.com/danroc/geoblock/pull/107))
- Update module github.com/go-playground/validator/v10 to v10.25.0 ([#106](https://github.com/danroc/geoblock/pull/106))
- Update alpine docker tag to v3.21.3 ([#105](https://github.com/danroc/geoblock/pull/105))
- Update golang docker tag to v1.24.0 ([#104](https://github.com/danroc/geoblock/pull/104))
- Update module golang.org/x/net to v0.35.0 ([#103](https://github.com/danroc/geoblock/pull/103))
- Update module golang.org/x/crypto to v0.33.0 ([#102](https://github.com/danroc/geoblock/pull/102))
- Update module golang.org/x/sys to v0.30.0 ([#99](https://github.com/danroc/geoblock/pull/99))
- Update module golang.org/x/text to v0.22.0 ([#100](https://github.com/danroc/geoblock/pull/100))
- Update golang docker tag to v1.23.6 ([#101](https://github.com/danroc/geoblock/pull/101))
- Update golang docker tag to v1.23.5 ([#98](https://github.com/danroc/geoblock/pull/98))
- Update module github.com/go-playground/validator/v10 to v10.24.0 ([#97](https://github.com/danroc/geoblock/pull/97))

## [0.1.16] - 2025-01-09

### Added

- Add basic e2e tests ([#70](https://github.com/danroc/geoblock/pull/70))
- Add a license ([#76](https://github.com/danroc/geoblock/pull/76))
- Add "Unreleased" to CHANGELOG ([#86](https://github.com/danroc/geoblock/pull/86))
- Document some of the main features ([#87](https://github.com/danroc/geoblock/pull/87))

### Changed

- Update README following addition of e2e tests ([#72](https://github.com/danroc/geoblock/pull/72))
- Refactor Makefile ([#73](https://github.com/danroc/geoblock/pull/73))
- Small improvements to Makefile ([#74](https://github.com/danroc/geoblock/pull/74))
- Rename some Makefile targets ([#75](https://github.com/danroc/geoblock/pull/75))
- Update module golang.org/x/crypto to 0.31.0 ([#77](https://github.com/danroc/geoblock/pull/77))
- Enable indirect packages updates ([#78](https://github.com/danroc/geoblock/pull/78))
- Update module github.com/gabriel-vasile/mimetype to v1.4.7 ([#79](https://github.com/danroc/geoblock/pull/79))
- Update module golang.org/x/net to v0.32.0 ([#80](https://github.com/danroc/geoblock/pull/80))
- Run `go mod tidy` on package update ([#81](https://github.com/danroc/geoblock/pull/81))
- Version dev dependencies ([#82](https://github.com/danroc/geoblock/pull/82))
- Rename targets ([#84](https://github.com/danroc/geoblock/pull/84))
- Move dev deps back to main Makefile ([#85](https://github.com/danroc/geoblock/pull/85))
- Update module golang.org/x/net to v0.33.0 ([#88](https://github.com/danroc/geoblock/pull/88))
- Small grammar fixes ([#89](https://github.com/danroc/geoblock/pull/89))
- Update module github.com/gabriel-vasile/mimetype to v1.4.8 ([#90](https://github.com/danroc/geoblock/pull/90))
- Update module golang.org/x/sys to v0.29.0 ([#91](https://github.com/danroc/geoblock/pull/91))
- Update outputs of examples ([#92](https://github.com/danroc/geoblock/pull/92))
- Update module golang.org/x/crypto to v0.32.0 ([#93](https://github.com/danroc/geoblock/pull/93))
- Update module golang.org/x/net to v0.34.0 ([#94](https://github.com/danroc/geoblock/pull/94))
- Update alpine docker tag to v3.21.2 ([#95](https://github.com/danroc/geoblock/pull/95))

### Removed

- Remove broken CI action ([#69](https://github.com/danroc/geoblock/pull/69))
- Don't log redundant error message ([#71](https://github.com/danroc/geoblock/pull/71))

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

[Unreleased]: https://github.com/danroc/geoblock/compare/0.1.18...HEAD
[0.1.18]: https://github.com/danroc/geoblock/compare/v0.1.17...0.1.18
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

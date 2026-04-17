# Changelog

All notable changes to `ginlogctx` will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.2] - 2026-04-17

### Added
- Out-of-the-box request ID generation inside `ginlogctx.Middleware`
- Documentation explaining when `ginlogctx` is a simpler alternative to OpenTelemetry and APM

### Changed
- Simplified consumer setup so Gin services no longer need to add `gin-contrib/requestid` separately
- Updated examples and tests to reflect the package-owned request ID flow
- Improved README positioning around request correlation, logs, and Datadog-style log tracking

## [0.1.1] - 2026-04-17

### Added
- Docker-based local test workflow with `make docker-test`
- Test coverage for request completion logs, scoped fields, cleanup, and concurrent requests
- Representative JSON log output in tests to show the final expected log shape

### Changed
- Simplified the built-in request completion log to avoid `file` and `func`
- Improved documentation around request log configuration and Docker-based testing
- Aligned tests with the default `gin-contrib/requestid` behavior and realistic request IDs

## [0.1.0] - 2026-04-17

### Added
- Initial `ginlogctx` release
- Logrus hook for request-scoped Gin fields
- Gin middleware for binding `request_id` and custom fields to logs
- Configurable request completion logging
- Extensible custom field resolvers for values such as `user_id` and `product_id`
- README, contribution guide, and package documentation

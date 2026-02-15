# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Community and open-source governance documents.

## [0.1.0] - 2026-02-15

### Added

- ZIP upload endpoint (`POST /upload`) to convert archive content into a Git repository.
- Health endpoint (`GET /health`) for service status checks.
- Repository deletion endpoint (`DELETE /delete`, compatible with `POST /delete`).
- Automatic repository creation via Gitea API when target repository does not exist.
- ZIP extraction behavior that trims a single common root folder.
- Filtering for common macOS/Windows junk files (`__MACOSX`, `.DS_Store`, `Thumbs.db`, `desktop.ini`, `._*`).

### Changed

- Returned `git_url` now prefers `GITEA_PUBLIC_URL` when configured.
- README and deployment docs aligned with runtime behavior and required environment variables.

### Fixed

- Added Traefik load balancer port label adjustments in deployment configuration.

### Security

- Added coordinated vulnerability reporting process via `SECURITY.md`.

### Known Limitations

- Upload size limit is `100MB`.
- Upload operation force-pushes to remote `main`.
- Deployment pipeline assumes repository visibility/accessibility from Coolify.

[Unreleased]: https://github.com/EagleFandel/zip-agent/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/EagleFandel/zip-agent/releases/tag/v0.1.0

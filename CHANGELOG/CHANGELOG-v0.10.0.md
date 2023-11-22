# Changelog


<a name="v0.10.0"></a>
## [v0.10.0] - 2023-11-22
### Features
- Add better http logging with more information
- Add compression to http server

### Fixes
- Fix Go version in prod Dockerfile
- Fix PWA status bar/theme colour for iOS
- Fix CI caching for go modules

### Refactoring
- Refactor tool install in CI to use go.mod version instead of hardcoded version
- Refactor tool versions to use the tools.go pattern instead of versions in the Justfile
- Refactor use wgo instead of concurrently for all watch tasks


[Unreleased]: https://github.com/RobinThrift/stuff/compare/v0.10.0...HEAD
[v0.10.0]: https://github.com/RobinThrift/stuff/compare/v0.9.2...v0.10.0

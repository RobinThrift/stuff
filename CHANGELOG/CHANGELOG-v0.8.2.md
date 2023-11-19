# Changelog


<a name="v0.8.2"></a>
## [v0.8.2] - 2023-11-19
### Features
- Add css minification to production build

### Fixes
- Fix potential null pointer panic when query fails
- Fix major issues with foreign key constraints being checked
- Fix double request to settings endpoint when changing theme
- Fix error with concurrent access to SQLite database in single requests

### Refactoring
- Refactor ad-hoc implementation with library code


[Unreleased]: https://github.com/RobinThrift/stuff/compare/v0.8.2...HEAD
[v0.8.2]: https://github.com/RobinThrift/stuff/compare/v0.8.1...v0.8.2

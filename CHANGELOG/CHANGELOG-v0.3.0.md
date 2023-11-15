# Changelog


<a name="v0.3.0"></a>
## [v0.3.0] - 2023-11-14
### Features
- Add theme and light/dark mode switcher
- Add new retro theme
- Add dependabot config
- Add autocomplete for location and position code
- Add autocomplete for custom attributes
- Add support for custom attributes
- Add proper shutdown handler to prevent db corruption
- Add simple file management for assets
- Add components and consumables import from Snipe-IT API
- Add asset type
- Add links to import assets page

### Fixes
- Fix manufacturers view having the wrong grouping clause
- Fix label sheet creator route
- Fix styles and colours to work a little better in light mode
- Fix nonsense autocomplete for value fields
- Fix incorrect fetching from DB in asset details page
- Fix showing single purchase info on when editing a consumable
- Fix formatting
- Fix missing initialisation
- Fix API errors
- Fix Avery L78710-20 template
- Fix duplicate categories in list command
- Fix not creating tags when importing
- Fix checkbox border colour in dark mode
- Fix dark mode

### Refactoring
- Refactor styles for better look and feel
- Refactor entire codebase to make more sense and be more consistent
- Refactor autocomplete to use browser native UI and builtins
- Refactor single asset view page
- Refactor to support multiple purchases for consumables

### Updates
- Update (deps-dev): Bump [@tailwindcss](https://github.com/tailwindcss)/forms in /frontend
- Update (deps): Bump golang.org/x/crypto from 0.14.0 to 0.15.0
- Update (deps): Bump github.com/microcosm-cc/bluemonday
- Update (deps-dev): Bump autoprefixer in /frontend
- Update (deps): Bump actions/setup-go from 3 to 4
- Update golangci-lint from 1.54.1 to 1.55.1


[Unreleased]: https://github.com/RobinThrift/stuff/compare/v0.3.0...HEAD
[v0.3.0]: https://github.com/RobinThrift/stuff/compare/v0.2.0...v0.3.0

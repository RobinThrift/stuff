# Changelog


<a name="v0.1.0"></a>
## v0.1.0 - 2023-10-09
### Features
- Add release workflow
- Add changelog generator
- Add minimal CI using GitHub actions
- Add docker build
- Add version info to prod build
- Add label sheet creation
- Add user management
- Add asset generation
- Add component system to go templates
- Add full CSV and JSON export for assets
- Add belongs to relationship to assets
- Add parts to assets
- Add sanitisation to missing fields
- Add command palette
- Add removal of illegal characters in SQLite full text search
- Add basic full text search
- Add basic tag list view
- Add autocomplete for categories when creating/editing an asset
- Add basic asset CRUD
- Add default currency config
- Add Alpine.js
- Add SQLite backed session store
- Add comments that explain which method belongs to which route
- Add better formatting and attribute printing to console log handler
- Add basic local login and initial setup
- Add custom type for SQLite date handling
- Add error page
- Add sensible default env vars for when running using just
- Add tailwind setup
- Add clean task
- Add linting
- Add main file
- Add basic server setup
- Add basic sqlite setup
- Add basic logging setup
- Add basic migration setup

### Fixes
- Fix linting errors
- Fix linting in CI
- Fix repo path and module name
- Fix formatting
- Fix biome tasks in root justfile
- Fix minor style bugs in edit/create asset page
- Fix icons auto rebuild
- Fix missing quotes for attribute
- Fix search params when using single field with no space
- Fix incorrect return type in ListCategories method
- Fix incorrect image url in asset view
- Fix incorrect formatting directives
- Fix missing copy task for js lib
- Fix autocomplete selection
- Fix typo
- Fix filename
- Fix formatting
- Fix formatting
- Fix not returning error

### Refactoring
- Refactor styles to add dark mode
- Refactor asset URLs to use either ID or tag
- Refactor navigation
- Refactor complex alpine components to external JS
- Refactor FTS to be able to search for specific fields
- Refactor justfile for better readability
- Refactor autocompleter to use server for better scalability
- Refactor static files
- Refactor migrator switching from sql-migrate to goose
- Refactor Assets nav link to link to /assets instead of /
- Refactor tool version in justfile to be centralised


[Unreleased]: https://github.com/RobinThrift/stuff/compare/v0.1.0...HEAD

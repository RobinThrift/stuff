export PATH := "./node_modules/.bin:" + env_var('PATH')
export STUFF_LOG_LEVEL := "debug"
export STUFF_LOG_FORMAT := "console"
export STUFF_ADDRESS := "localhost:8080"
export STUFF_AUTH_LOCAL_INITIAL_ADMIN_PASSWORD := "admin"
export STUFF_FILE_DIR := "files_dev_run"

staticcheck_version := "2023.1.5"
golangci_lint_version := "v1.54.2"
sql_migrate_version := "v1.5.2"
bobgen_version := "v0.22.0"
templ_version := "v0.2.316"
wgo_version := "v0.5.3"

sql_migrate_config := "./storage/database/sqlite/sqlmigrate.yaml"

_default:
    @just --list

fmt:
    go fmt ./...

lint:
	go run honnef.co/go/tools/cmd/staticcheck@{{staticcheck_version}} ./...
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@{{golangci_lint_version}} run ./...

run: _gen-templ _copy-js-libs icons
    go run ./bin/stuff

build: _gen-templ _copy-js-libs styles icons
    go build ./bin/stuff

styles:
    postcss ./views/styles.css -o ./build/styles.css --no-map

icons:
    rm -f build/*.svg
    svg-sprite \
        --symbol --symbol-dest="" \
        --symbol-prefix=".icon-%s" --symbol-sprite=icons.svg \
        --dest=build views/icons/*.svg

watch:
    concurrently "just _watch-go" "just _watch-styles" "just _watch-icons"

_watch-go: _copy-js-libs
    go run github.com/bokwoon95/wgo@{{wgo_version}} \
        -xfile '.*_templ.go' \
        -file '.*\.go' \
        -file '.*\.templ' \
        just _run-watch

_run-watch: _gen-templ
    go run -tags dev ./bin/stuff

_watch-styles:
    postcss ./views/styles.css -o ./build/styles.css --watch

_watch-icons:
    go run github.com/bokwoon95/wgo@{{wgo_version}} \
        -file 'views\/icons\/.*\.svg' \
        just icons

new-migration name:
    go run github.com/rubenv/sql-migrate/sql-migrate/...@{{sql_migrate_version}} new -env production -config={{sql_migrate_config}} {{name}}

alias gen := generate
generate:
    rm -f _stuff.db
    go run github.com/rubenv/sql-migrate/sql-migrate/...@{{sql_migrate_version}} up -env production -config={{sql_migrate_config}}
    go run github.com/stephenafamo/bob/gen/bobgen-sqlite@{{bobgen_version}} -c ./storage/database/sqlite/bob.yaml
    go fmt ./...
    rm _stuff.db
    just _gen-templ

_gen-templ:
    go run github.com/a-h/templ/cmd/templ@{{templ_version}} generate -path .

_copy-js-libs:
    cp node_modules/alpinejs/dist/cdn.min.js build/alpine.min.js
    cp node_modules/flatpickr/dist/flatpickr.min.js build/flatpickr.min.js
    cp node_modules/htmx.org/dist/htmx.min.js build/htmx.min.js
    cp node_modules/quick-score/dist/quick-score.esm.min.js build/quick-score.min.js

clean:
    rm -rf node_modules stuff.db build
    go clean -cache


export PATH := "./static/node_modules/.bin:" + env_var('PATH')
export STUFF_LOG_LEVEL := "debug"
export STUFF_LOG_FORMAT := "console"
export STUFF_ADDRESS := "localhost:8080"
export STUFF_AUTH_LOCAL_INITIAL_ADMIN_PASSWORD := "admin"
export STUFF_FILE_DIR := ".run/files_dev_run"
export STUFF_DATABASE_PATH := ".run/stuff.db"

staticcheck_version := "2023.1.5"
golangci_lint_version := "v1.54.2"
goose_version := "v3.15.0"
bobgen_version := "v0.22.0"
templ_version := "v0.2.316"
wgo_version := "v0.5.3"

_default:
    @just --list

fmt:
    go fmt ./...

lint:
	go run honnef.co/go/tools/cmd/staticcheck@{{staticcheck_version}} ./...
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@{{golangci_lint_version}} run ./...

test flags="-failfast -v -timeout 5m":
    @[ -d static/build ] || (mkdir static/build && touch static/build/styles.css)
    go test {{ flags }} ./...

run: _gen-templ _copy-js-libs icons
    mkdir -p .run
    go run ./bin/stuff

build: _gen-templ _copy-js-libs styles icons
    go build ./bin/stuff

styles: _npm-install
    postcss -c static/postcss.config.js ./static/src/styles.css -o ./static/build/styles.css --no-map

icons: _npm-install
    rm -f staic/build/*.svg
    svg-sprite \
        --symbol --symbol-dest="" \
        --symbol-prefix=".icon-%s" --symbol-sprite=icons.svg \
        --dest=static/build static/src/icons/*.svg

watch: _npm-install
    concurrently "just _watch-go" "just _watch-styles" "just _watch-icons"

_watch-go: _copy-js-libs
    go run github.com/bokwoon95/wgo@{{wgo_version}} \
        -xfile '.*_templ.go' \
        -file '.*\.go' \
        -file '.*\.templ' \
        just _run-watch

_run-watch: _gen-templ
    mkdir -p .run
    go run -tags dev ./bin/stuff

_watch-styles:
    postcss ./static/src/styles.css -o ./static/build/styles.css --watch

_watch-icons:
    go run github.com/bokwoon95/wgo@{{wgo_version}} \
        -file 'static/src/\/icons\/.*\.svg' \
        just icons

new-migration name:
    @rm -f _stuff.db
    go run github.com/pressly/goose/v3/cmd/goose@{{goose_version}} -table migrations -dir storage/database/sqlite/migrations sqlite3 ./_stuff.db create {{name}} sql
    @rm -f _stuff.db

alias gen := generate
generate:
    @rm -f _stuff.db
    go run github.com/pressly/goose/v3/cmd/goose@{{goose_version}} -table migrations -dir storage/database/sqlite/migrations sqlite3 ./_stuff.db up
    go run github.com/stephenafamo/bob/gen/bobgen-sqlite@{{bobgen_version}} -c ./storage/database/sqlite/bob.yaml
    go fmt ./...
    @rm _stuff.db
    just _gen-templ

_gen-templ:
    go run github.com/a-h/templ/cmd/templ@{{templ_version}} generate -path .

_copy-js-libs: _npm-install
    -mkdir static/build
    cp static/node_modules/alpinejs/dist/cdn.min.js static/build/alpine.min.js
    cp static/node_modules/flatpickr/dist/flatpickr.min.js static/build/flatpickr.min.js
    cp static/node_modules/htmx.org/dist/htmx.min.js static/build/htmx.min.js
    cp static/node_modules/quick-score/dist/quick-score.esm.min.js static/build/quick-score.min.js

_npm-install:
    [ -d static/node_modules ] || (cd static && npm i --no-audit --no-fund)

clean:
    rm -rf node_modules build .run
    go clean -cache


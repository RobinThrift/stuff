export PATH := "./node_modules/.bin:" + env_var('PATH')
export STUFF_LOG_LEVEL := "debug"
export STUFF_LOG_FORMAT := "console"
export STUFF_ADDRESS := "localhost:8080"
export STUFF_AUTH_LOCAL_INITIAL_ADMIN_PASSWORD := "admin"

staticcheck_version := "2023.1.5"
golangci_lint_version := "v1.54.2"
sql_migrate_version := "v1.5.2"
bobgen_version := "v0.22.0"
templ_version := "v0.2.316"
wgo_version := "v0.5.3"

sql_migrate_config := "./storage/database/sqlite/sqlmigrate.yaml"

_default:
    @just --list

clean:
    rm -rf node_modules stuff.db build
    go clean -cache

fmt:
    go fmt ./...

lint:
	go run honnef.co/go/tools/cmd/staticcheck@{{staticcheck_version}} ./...
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@{{golangci_lint_version}} run ./...

build: _gen-templ tailwind-build _copy-js-libs
    go build ./bin/stuff

run: _gen-templ tailwind-build _copy-js-libs
    go run ./bin/stuff

watch: _copy-js-libs
    go run github.com/bokwoon95/wgo@{{wgo_version}} \
        -xdir node_modules \
        -xdir build \
        -xfile '.*_templ.go' \
        -xfile 'justfile' \
        -xfile 'stuff.db' \
        just _watch-run

_watch-run: _gen-templ
    go run -tags dev ./bin/stuff

tailwind-build:
    tailwindcss -i ./views/styles.css -o ./build/styles.css

alias tw := tailwind
tailwind:
    tailwindcss -i ./views/styles.css -o ./build/styles.css --watch


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

new-migration name:
    go run github.com/rubenv/sql-migrate/sql-migrate/...@{{sql_migrate_version}} new -env production -config={{sql_migrate_config}} {{name}}

_copy-js-libs:
    cp node_modules/alpinejs/dist/cdn.min.js build/alpine.min.js

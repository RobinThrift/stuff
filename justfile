export PATH := "./node_modules/.bin:" + env_var('PATH')

sql_migrate_config := "./storage/database/sqlite/sqlmigrate.yaml"

_default:
    @just --list

fmt:
    go fmt ./...
    go run github.com/a-h/templ/cmd/templ@0.2.316 fmt .

alias gen := generate
generate:
    rm -f stuff.db
    go run github.com/rubenv/sql-migrate/sql-migrate/...@v1.5.2 up -env production -config={{sql_migrate_config}}
    SQLITE_DSN=stuff.db SQLITE_OUTPUT="storage/database/sqlite/models" SQLITE_PKGNAME="models" go run github.com/stephenafamo/bob/gen/bobgen-sqlite@v0.22.0
    go run github.com/a-h/templ/cmd/templ@v0.2.316 generate
    rm stuff.db

new-migration name:
    go run github.com/rubenv/sql-migrate/sql-migrate/...@v1.5.2 new -env production -config={{sql_migrate_config}} {{name}}

version        := env_var_or_default("VERSION", "dev")
go_ldflgas     := env_var_or_default("GO_LDFLGAS", "") + " -X 'github.com/RobinThrift/stuff.Version=" + version + "'"
go_build_flags := env_var_or_default("GO_BUILD_FLAGS", "")
oci_repo       := env_var_or_default("OCI_REPO", "ghcr.io/robinthrift/stuff")


export PATH := "./static/node_modules/.bin:" + "./node_modules/.bin:" + env_var('PATH')
export STUFF_LOG_LEVEL := "debug"
export STUFF_LOG_FORMAT := "console"
export STUFF_ADDRESS := "localhost:8080"
export STUFF_AUTH_LOCAL_INITIAL_ADMIN_PASSWORD := "admin"
export STUFF_FILE_DIR := ".run/files_dev_run"
export STUFF_DATABASE_PATH := ".run/stuff.db"
export STUFF_USE_SECURE_COOKIES := "false"

_default:
    @just --list

fmt: _npm-install
    go fmt ./...
    cd static && biome format --write src/*.ts

lint: _npm-install
    @[ -d static/build ] || (mkdir static/build && touch static/build/styles.css)
    staticcheck ./...
    golangci-lint run ./...
    cd static && biome check src/*.ts

test *flags="-failfast -v -timeout 5m":
    @[ -d static/build ] || (mkdir static/build && touch static/build/styles.css)
    go test {{ flags }} ./...

run: build-js build-icons _fonts
    mkdir -p .run
    go run -tags dev ./bin/stuff


watch: _npm-install _fonts
    mkdir -p .run
    concurrently "just _watch-go" "just _watch-styles" "just _watch-icons" "just _watch-js"

_watch-go:
    wgo \
        -file '.*\.go' \
        -xfile '.*_test\.go' \
        go run -tags dev ./bin/stuff

_watch-styles:
    postcss ./static/src/styles.css -o ./static/build/styles.css --watch

_watch-js:
    cd static && esbuild src/index.ts --format=esm --target=es2020 --bundle --outfile=build/bundle.min.js --watch

_watch-icons:
    wgo \
        -file 'static/src/icons/.*\.svg' \
        just build-icons

build: build-js build-js build-styles build-icons _fonts
    just _build-go

_build-go:
    go build -ldflags="{{go_ldflgas}}" {{ go_build_flags }} -o build/stuff ./bin/stuff

build-styles: _npm-install
    postcss -c static/postcss.config.js ./static/src/styles.css -o ./static/build/styles.css --no-map

build-js: _npm-install
    cd static && esbuild src/index.ts --format=esm --target=es2020 --minify --bundle --outfile=build/bundle.min.js

build-icons: _npm-install
    rm -f staic/build/*.svg
    svg-sprite \
        --symbol --symbol-dest="" \
        --symbol-prefix=".icon-%s" --symbol-sprite=icons.svg \
        --dest=static/build static/src/icons/*.svg

docker_cmd := env_var_or_default("DOCKER_CMD", "build")
build-oci-image:
    docker {{ docker_cmd }} -f ./deployment/Dockerfile  -t {{ oci_repo }}:{{ version }} .

run-oci-image: build-oci-image
    docker run --rm \
        -e STUFF_LOG_LEVEL={{ STUFF_LOG_LEVEL }} \
        -e STUFF_LOG_FORMAT={{ STUFF_LOG_FORMAT }} \
        -e STUFF_AUTH_LOCAL_INITIAL_ADMIN_PASSWORD={{ STUFF_AUTH_LOCAL_INITIAL_ADMIN_PASSWORD }} \
        -e STUFF_USE_SECURE_COOKIES={{ STUFF_USE_SECURE_COOKIES }} \
        -p 8080:8080 \
        {{ oci_repo }}:{{ version }}

# generate a release with the given tag
release tag:
    just changelog {{tag}}
    git add CHANGELOG
    git commit -m "Releasing version {{tag}}"
    git tag {{tag}}
    git push
    git push origin {{tag}}

# generate a changelog using github.com/git-chglog/git-chglog
changelog tag: _go-tools
    git-chglog -o CHANGELOG/CHANGELOG-{{tag}}.md --next-tag {{tag}} --sort semver --config CHANGELOG/config.yml {{ tag }}
    echo "- [CHANGELOG-{{tag}}.md](./CHANGELOG-{{tag}}.md)" >> CHANGELOG/README.md

new-migration name: _go-tools
    @rm -f _stuff.db
    goose -table migrations -dir storage/database/sqlite/migrations sqlite3 ./_stuff.db create {{name}} sql
    @rm -f _stuff.db

alias gen := generate
generate: _go-tools
    @rm -f _stuff.db
    goose -table migrations -dir storage/database/sqlite/migrations sqlite3 ./_stuff.db up
    bobgen-sqlite -c ./storage/database/sqlite/bob.yaml
    go fmt ./...
    @rm _stuff.db

install:
    just _npm-install
    just _go-tools

_npm-install:
    [ -d static/node_modules ] || (cd static && npm i --no-audit --no-fund)

_fonts:
    [ -f static/build/fonts/OpenSans-Regular.ttf ] || (mkdir -p static/build/fonts && curl -L https://github.com/googlefonts/opensans/raw/main/fonts/ttf/OpenSans-Regular.ttf -o static/build/fonts/OpenSans-Regular.ttf)

staticcheck_version := "2023.1.5"
golangci_lint_version := "v1.54.2"
goose_version := "v3.15.0"
bobgen_version := "v0.22.0"
wgo_version := "v0.5.3"
git_chglog_version := "v0.15.4"
_go-tools:
    @if ! type -p wgo > /dev/null ; then go install github.com/bokwoon95/wgo@{{wgo_version}} ; fi
    @if ! type -p staticcheck > /dev/null ; then go install honnef.co/go/tools/cmd/staticcheck@{{staticcheck_version}} ; fi
    @if ! type -p golangci-lint > /dev/null ; then go install github.com/golangci/golangci-lint/cmd/golangci-lint@{{golangci_lint_version}} ; fi
    @if ! type -p goose > /dev/null ; then go install github.com/pressly/goose/v3/cmd/goose@{{goose_version}} ; fi
    @if ! type -p goose > /dev/null ; then go install github.com/pressly/goose/v3/cmd/goose@{{goose_version}} ; fi
    @if ! type -p bobgen-sqlite > /dev/null ; then go install github.com/stephenafamo/bob/gen/bobgen-sqlite@{{bobgen_version}} ; fi
    @if ! type -p git-chglog > /dev/null ; then go install github.com/git-chglog/git-chglog/cmd/git-chglog@{{bobgen_version}} ; fi


clean:
    rm -rf node_modules build .run
    go clean -cache


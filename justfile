_default:
    @just --list

fmt:
    go fmt ./...
    go run github.com/a-h/templ/cmd/templ@0.2.316 fmt .

# List available recipes
_default:
    @just --justfile {{justfile()}} --list --unsorted

# Run tests and open coverage report
coverage:
    go test -covermode=count -coverpkg=./... -coverprofile cover.out -v ./...
    go tool cover -html cover.out -o cover.html
    xdg-open cover.html

# Run local pkgsite server for package docs preview
pkgsite:
    pkgsite -http localhost:3948 -open
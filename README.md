# Continuum TMDB Plugin

First-party Continuum metadata plugin backed by TMDB.

## Dependency Model

This repository consumes `github.com/ContinuumApp/continuum-plugin-sdk` as a normal Go module dependency. CI and release builds run with `GOWORK=off` and expect the SDK version in `go.mod` to resolve from a published semver tag.

For local multi-repo development, use a temporary `replace` or a local `go.work` that points at `dev/github/continuum-plugin-sdk`. Do not commit machine-local filesystem replaces as the supported release path.

## Development

```sh
go test ./...
go build .
```

## License

`continuum-plugin-tmdb` is licensed under `AGPL-3.0-or-later`. See [LICENSE](LICENSE).

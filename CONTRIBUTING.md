# Contributing

Thanks for helping out on **nx-cache-service** — the self-hosted remote cache server for Nx.

This document covers the basics: how to get the service running, how to verify your change, and how to get it reviewed
and merged. For what the service does and how it is configured, see [`README.md`](./README.md) — endpoints, environment
variables and the Nx client setup all live there.

## Table of contents

- [Getting started](#getting-started)
- [Code style](#code-style)
- [Verifying your change](#verifying-your-change)
- [Pull request or issue?](#pull-request-or-issue)
- [Pull request process](#pull-request-process)
  - [Commit messages](#commit-messages)

## Getting started

### Prerequisites

- **Go 1.25 or newer.** The version is pinned in [`go.mod`](./go.mod) and CI installs it from there, so keep the two in
  step if you bump it.
- **Docker** with Compose, for the local S3 backend and for building the container image.
- Nothing else. `gofmt` and `go vet` ship with Go, and the only third-party dependency is the AWS SDK.

### Run the service

```sh
# 1. build the binary
make build

# 2. run it against the local disk store — no object storage needed
CI_TOKEN=dev-ci-token make run

# 3. check it is up
curl http://localhost:8080/health
```

`CI_TOKEN` is required; the service exits at startup without it. To work against S3 instead, bring up the Compose
stack, which runs the service with `CACHE_STORE=s3` against a local SeaweedFS bucket:

```sh
docker compose up --build
```

### Where things live

| Path                     | What it is                                                                        |
| ------------------------ | --------------------------------------------------------------------------------- |
| `cmd/nx-cache-service/`  | Entrypoint: token setup, store selection, TLS, HTTP server                        |
| `internal/api/`          | Routes, handlers, auth middleware, request logging, panic recovery, JSON helpers  |
| `internal/auth/`         | Static bearer token authenticator                                                 |
| `internal/cache/`        | The `Store` interface and its disk and S3 implementations                         |
| `internal/env/`          | Environment variable keys                                                         |
| `docker/`                | SeaweedFS S3 credentials used by Compose                                          |
| `.github/workflows/`     | CI pipeline                                                                       |

### Making a change

1. Branch off `master` with a short, descriptive name: `git checkout -b e2e-testing`.
2. Write the code following the [conventions below](#code-style).
3. Update [`README.md`](./README.md) whenever you change something a user can observe — an endpoint, a status code, an
   environment variable, a default.
4. Run `make check` before pushing.

## Code style

- **No comments.** The Go source contains none, including doc comments. Name things so the code reads without them; if
  a block needs explaining, that is a signal to refactor it.
- **Standard library first.** Routing is `net/http`'s `ServeMux` with method-and-pattern routes; logging is `log/slog`.
  The AWS SDK is the only third-party dependency — keep it that way unless there is a strong reason.
- **Handlers are constructors that return `http.Handler`**, taking their dependencies as arguments:
  `func handleCacheGet(store cache.Store, log *slog.Logger) http.Handler`. Register them in `addRoutes`.
- **Storage backends implement `cache.Store`.** Add a new one in `internal/cache`, then wire it into `openStore` in
  `cmd/nx-cache-service/main.go`.
- **Stores return sentinel errors** (`cache.ErrNotFound`, `cache.ErrExists`); the API layer maps them to status codes.
  Keep HTTP concerns out of the store and storage concerns out of the handler.
- **All responses go through `writeJSON` or `writeErrorResponse`** so the JSON error shape stays consistent.
- **Every path parameter goes through `validatePathParam`** before it reaches a store — that is what keeps `..` and
  path separators out of object keys and file paths.
- **New environment variables are `EnvKey` constants in `internal/env`**, read with `GetValue()`, never `os.Getenv`
  at the call site.
- **Formatting is `gofmt`.** `make check` runs `gofmt -l` and `go vet ./...` and fails on either.

## Verifying your change

There is no automated test suite in this repository yet, so verification is `make check` plus exercising the change by
hand. Describe what you ran in the pull request.

```sh
make check   # gofmt -l and go vet ./...
make build   # compile
```

Exercise the API against a running server — both stores go through the same handlers, so check the one your change
touches:

```sh
printf 'hello-artifact' | curl -X PUT --data-binary @- \
  -H 'Authorization: Bearer dev-ci-token' \
  http://localhost:8080/project/my-app/v1/cache/abc123

curl -H 'Authorization: Bearer dev-ci-token' \
  http://localhost:8080/project/my-app/v1/cache/abc123
```

Worth walking through for anything touching auth, validation or the stores: a cache miss returns `404`, a repeated
`PUT` of the same hash returns `409`, a write with `READ_ONLY_TOKEN` returns `403`, a request with no token returns
`401`, and an invalid `project` or `hash` returns `400`.

If you changed the Dockerfile, the Compose stack or anything the container depends on, confirm the image still builds
and starts:

```sh
docker build -t nx-cache-service .
docker compose up --build
```

## Pull request or issue?

Small, self-contained changes go straight to a pull request. Anything large enough that the approach itself is worth
discussing starts as an issue, so the direction is agreed before you spend time on the code.

**Open a pull request directly** for:

- a bug fix
- a small performance improvement that does not require a big refactor of existing code
- a small feature

Explain in the description **what** the change does, **why** it is needed, and **what the impact is** — behaviour that
changes, new or renamed configuration, anything a user of the service would notice.

**Open a GitHub issue first** for:

- a large feature
- a rewrite or big refactor of existing code
- a change to the HTTP API, the storage backends or the configuration surface that existing deployments would have to
  adapt to

Describe the problem you are solving and the approach you have in mind. Once a maintainer agrees on the direction, go
ahead and implement it and open a pull request that references the issue. You are welcome to implement it yourself —
just get the approval first, so the work is not wasted on an approach that will not be merged.

Issues already labelled **`contributions welcome`** are the exception: the direction on those is settled, so you can
pick one up and open a pull request straight away, without waiting for approval. Leave a comment on the issue before
you start so nobody duplicates the work.

If you are not sure which side of the line a change falls on, open an issue. It is cheaper than a rejected pull
request.

## Pull request process

1. **Open the PR against `master`**, covering the what, why and impact described above, plus how you verified it. Link
   the issue if the change started as one.
2. **Keep it focused.** One logical change per pull request, and one logical change per commit — the history in this
   repository is deliberately granular, and it makes review and `git bisect` easier.
3. **Make sure CI is green.** [`.github/workflows/ci.yml`](./.github/workflows/ci.yml) runs on every pull request: a
   `build` job running `make check` and `make build`, then a `docker` job that builds the image with Buildx. Image
   publishing to GHCR is currently disabled, so the docker job only proves the image builds.
4. **Address feedback with follow-up commits** rather than force-pushing over the review, so reviewers can see what
   changed since they last looked.

### Commit messages

Commit messages are a **single-line subject with no body**. The diff shows what changed; the subject says what it
achieves.

- Imperative mood, capitalised, no trailing period, ideally under ~72 characters.
- One commit per logical change — split unrelated work rather than batching it.
- No trailers, prefixes or ticket references.

Examples from the history:

```
Implement disk-backed Get and Put for the cache store
Validate project/hash path params to prevent path traversal
Return 500 from the cache get handler when the store fails
Add CI workflow that builds the binary and publishes container images
```

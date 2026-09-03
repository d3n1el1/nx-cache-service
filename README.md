# nx-cache-service

A self-hosted remote cache server for [Nx](https://nx.dev), written in Go.

It implements the Nx self-hosted remote cache HTTP protocol (`GET`/`PUT /v1/cache/{hash}` with bearer
token auth), namespaced per project, and stores artifacts in any S3-compatible bucket. A local disk
store is included for development.

## Features

- **Nx-compatible cache protocol** — `GET`/`PUT` cache endpoints plus a `GET /health` probe.
- **Per-project namespacing** — cache routes are scoped under `/project/{project}`, so one server can
  serve many workspaces without hash collisions.
- **S3-backed storage** — any S3-compatible object storage (AWS, MinIO, SeaweedFS, R2, …). A local
  disk store is also available for development, but is not meant for production.
- **Immutable artifacts** — writing a hash that already exists returns `409 Conflict` instead of
  overwriting it.
- **Read-only tokens** — a second token for developers' local setups and for PR and other
  non-main-branch builds: they read from the cache, while only the main branch build writes to it.
  Every request needs a token — the cache is never readable anonymously.
- **Constant-time token comparison** — tokens are compared with `crypto/subtle`.
- **Optional TLS** — serve HTTPS directly by pointing the service at a cert and key.
- **Structured logging and panic recovery** — every request is logged via `log/slog`; a panicking
  handler returns `500` instead of dropping the connection.

## Quick start

Run the service in Docker, backed by an S3 bucket:

```bash
docker run -d --name nx-cache-service \
  -p 8080:8080 \
  -e CI_TOKEN=replace-with-a-strong-token \
  -e READ_ONLY_TOKEN=replace-with-another-token \
  -e CACHE_STORE=s3 \
  -e S3_BUCKET_NAME=my-nx-cache \
  -e S3_REGION=eu-central-1 \
  -e S3_ACCESS_KEY_ID=<your-access-key-id> \
  -e S3_SECRET_ACCESS_KEY=<your-secret-access-key> \
  ghcr.io/d3n1el1/nx-cache-service:latest
```

Images are published to `ghcr.io/d3n1el1/nx-cache-service` on every push to `master`. `latest` follows
the newest `master` build; every build is also tagged with its commit SHA, so pin
`ghcr.io/d3n1el1/nx-cache-service:<commit-sha>` in production and upgrade deliberately. The images are
built for `linux/amd64` and `linux/arm64`.

Drop `S3_ACCESS_KEY_ID` and `S3_SECRET_ACCESS_KEY` when the host already provides credentials (instance
role, `AWS_*` environment variables, mounted `~/.aws/config`) — the AWS SDK's default credential chain
is used when they are unset. Add `-e S3_ENDPOINT_URL=https://…` for non-AWS S3-compatible storage such
as MinIO or Cloudflare R2.

Check that it is up:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

Store and read back an artifact:

```bash
printf 'hello-artifact' | curl -X PUT --data-binary @- \
  -H 'Authorization: Bearer replace-with-a-strong-token' \
  http://localhost:8080/project/my-app/v1/cache/abc123

curl -H 'Authorization: Bearer replace-with-a-strong-token' \
  http://localhost:8080/project/my-app/v1/cache/abc123
# hello-artifact
```

## Configuration

All configuration is read from environment variables.

### Server

| Variable          | Required | Default          | Description                                                                 |
| ----------------- | -------- | ---------------- | --------------------------------------------------------------------------- |
| `CI_TOKEN`        | yes      | —                | Bearer token with read **and** write access. The service refuses to start without it. |
| `READ_ONLY_TOKEN` | no       | —                | Second bearer token with read-only access. Write requests with it return `403`. Leave it unset only if `CI_TOKEN` is the sole token in use — there is no anonymous read, so without it nobody can read the cache without also being able to write to it. |
| `ADDR`            | no       | `localhost:8080` | Listen address. The Docker image sets `:8080` so the container is reachable from outside. |
| `CACHE_STORE`     | no       | `disk`           | Storage backend. Use `s3` in production. The default, `disk`, writes artifacts to `./.nx-cache-service` and is for local development only. |

### TLS

| Variable         | Required | Default | Description                     |
| ---------------- | -------- | ------- | ------------------------------- |
| `TLS_CERT_FILE`  | no       | —       | Path to the TLS certificate file. |
| `TLS_KEY_FILE`   | no       | —       | Path to the TLS private key file. |

Set both to serve HTTPS (TLS 1.2 minimum), or neither to serve plain HTTP. Setting only one is a
startup error.

### S3 store (`CACHE_STORE=s3`)

| Variable               | Required | Description                                                                                   |
| ---------------------- | -------- | --------------------------------------------------------------------------------------------- |
| `S3_BUCKET_NAME`       | yes      | Bucket that holds the cached artifacts.                                                        |
| `S3_REGION`            | no       | AWS region. Falls back to the default AWS config chain when unset.                             |
| `S3_ACCESS_KEY_ID`     | no       | Static access key. Must be set together with `S3_SECRET_ACCESS_KEY`.                           |
| `S3_SECRET_ACCESS_KEY` | no       | Static secret key. Must be set together with `S3_ACCESS_KEY_ID`.                               |
| `S3_ENDPOINT_URL`      | no       | Custom endpoint for S3-compatible storage (MinIO, SeaweedFS, R2). Enables path-style addressing. |

## Usage with Nx

Point your Nx workspace at the server. Nx appends `/v1/cache/{hash}` to the configured server URL, so
the URL must include the `/project/{project}` prefix this service serves cache routes under:

```bash
NX_SELF_HOSTED_REMOTE_CACHE_SERVER="http://localhost:8080/project/my-app"
NX_SELF_HOSTED_REMOTE_CACHE_ACCESS_TOKEN="replace-with-a-strong-token"
```

Give `CI_TOKEN` to the main branch build only — that is the one job that should populate the cache.
Everyone else gets `READ_ONLY_TOKEN`: developers running builds on their own machines, and PR and
other branch builds. They get cache hits from the main branch's artifacts without being able to write
anything into the cache themselves.

Configure `READ_ONLY_TOKEN` if anyone other than the main branch build is meant to use the cache.
Requests without a recognised token are rejected with `401`, so if it is left unset the only way to
give developers and PR builds cache hits is to hand them `CI_TOKEN` — which also lets them write.

## API

| Method   | Path                                   | Auth              | Description                                     |
| -------- | -------------------------------------- | ----------------- | ----------------------------------------------- |
| `GET`    | `/project/{project}/v1/cache/{hash}`   | read (either token) | Download a cached artifact.                     |
| `PUT`    | `/project/{project}/v1/cache/{hash}`   | write (`CI_TOKEN`)  | Upload an artifact. Existing hashes are never overwritten. |
| `DELETE` | `/project/{project}/flush`             | write (`CI_TOKEN`)  | Delete every artifact belonging to a project.   |
| `GET`    | `/health`                              | none              | Liveness probe. Returns `{"status":"ok"}`.      |

`project` and `hash` must match `^[A-Za-z0-9._-]+$` and must not be `.` or `..`.

A successful `GET` responds with `Content-Type: application/octet-stream` and a `Content-Length`
header. A successful `PUT` or `DELETE` responds with `{}`.

### Status codes

| Code  | Meaning                                                                      |
| ----- | ---------------------------------------------------------------------------- |
| `400` | `project` or `hash` is empty or contains invalid characters.                 |
| `401` | Missing or unrecognised bearer token. Responds with `WWW-Authenticate: Bearer`. |
| `403` | Known token without permission for the action (e.g. `READ_ONLY_TOKEN` on `PUT`). |
| `404` | The artifact is not in the cache.                                            |
| `409` | The artifact already exists — artifacts are immutable.                       |
| `500` | The storage backend failed.                                                  |

Errors are returned as JSON:

```json
{ "status": 404, "reason": "Not Found" }
```

## Development

Requires Go 1.25+ (see `go.mod`).

### Run from source

The disk store is the default, so no object storage is needed. It is a development convenience only —
deploy with `CACHE_STORE=s3`:

```bash
CI_TOKEN=dev-ci-token make run
```

The server listens on `localhost:8080` and writes artifacts to `./.nx-cache-service`.

### Run against a local S3

Docker Compose brings up the service together with
[SeaweedFS](https://github.com/seaweedfs/seaweedfs) as a local S3 backend and creates the `nx-cache`
bucket for you. SeaweedFS is for local development only — use a real bucket in production:

```bash
docker compose up --build
```

This runs the service with `CACHE_STORE=s3` against SeaweedFS on `http://seaweedfs:8333`, exposes it
on `http://localhost:8080`, and defaults `CI_TOKEN` to `dev-ci-token`. Override the tokens from your
environment:

```bash
CI_TOKEN=my-token READ_ONLY_TOKEN=my-ro-token docker compose up --build
```

### Make targets

```bash
make build      # build bin/nx-cache-service
make run        # go run ./cmd/nx-cache-service
make fmt        # gofmt -w .
make fmt-check  # fail if anything is not gofmt-ed
make vet        # go vet ./...
make check      # fmt-check + vet
make tidy       # go mod tidy
make all        # check + build
make clean      # remove bin/
```

CI (`.github/workflows/ci.yml`) runs `make check` and `make build` on every push to `master` and on
pull requests, then builds the Docker image with Buildx. Only pushes to `master` publish it, to
`ghcr.io/d3n1el1/nx-cache-service` as `latest` and `<commit-sha>`; pull requests build the image
without pushing it.

See [`CONTRIBUTING.md`](./CONTRIBUTING.md) for code style, how to verify a change, and the pull request process.

## Project layout

```
cmd/nx-cache-service/   entrypoint: config, store selection, TLS, HTTP server
internal/api/           routes, handlers, auth middleware, logging, JSON helpers
internal/auth/          static bearer token authenticator
internal/cache/         Store interface with disk and S3 implementations
internal/env/           environment variable keys
docker/                 SeaweedFS S3 credentials for Compose
```

# exchange-rates

A Go microservice that fetches currency exchange rates from the Bank of Latvia
ECB RSS feed and serves them over HTTP.

- `cli serve` — runs the HTTP API.
- `cli fetch` — pulls current rates for 10 preselected currencies (one
  goroutine per currency) and persists them to MySQL.

I tried to design packages around features, not types like many people do. This has many benefits like:
* as many implementation details are hidden as possible
* it's easier to understand what the package really does (no unneeded types exported)
* easier to understand dependencies of reach package

... but it has costs like it's much harder to keep the API clean on the package level. As we all know, we are all lazy :)

## Endpoints

| Method | Path                                  | Description                          |
|--------|---------------------------------------|--------------------------------------|
| GET    | `/api/v1/rates/latest`                | Latest known rate per currency.      |
| GET    | `/api/v1/rates/history/{currency}`    | History for one currency (DESC).     |

At first, rates will be empty. You have to run the `fetch` command to populate the DB.

## Quick start

Requires Docker and Compose v2.

```bash
# 1. Bring up MySQL, run migrations, and start the server.
docker compose up --build -d

# 2. In another shell, populate the database with fresh rates.
docker compose run --rm fetch

# 3. Hit the API.
curl http://localhost:8080/api/v1/rates/latest
curl http://localhost:8080/api/v1/rates/history/USD
```

## Testing

```bash
# Unit tests (fast, no infra):
go test ./...

# End-to-end tests (real MySQL via testcontainers, requires Docker):
go test -tags=e2e -count=1 ./e2e/...
```
## Observability

The service exports OTLP/gRPC to a collector you run yourself. From `docker compose`, that endpoint resolves to `host.docker.internal:4317`; running on the host, it's `localhost:4317`. Override with `OTEL_EXPORTER_OTLP_ENDPOINT`.

When you set up the endpoint, you'll get traces, logs and metrics.

# Notes
The API is pretty simple, so `net/http` was the library of choice as `Echo`,
`Gin` and other equivalent libraries were not mentioned in the task
and position descriptions.

`PgBouncer` was not used here due to the limited scale of this service.
Connections are pooled by the database driver instead (`pgx`).

Migrations are handled by `golang-migrate/migrate`.

TLS is not handled here since it would usually be done by a reverse proxy like
Caddy or Nginx in production which seems to be out of the assignment scope.

The API endpoints are mostly as requested in the assignment:
1. `POST /currency/add/{coin_symbol}` - add to watchlist.
2. `POST /currency/remove/{coin_symbol}` - remove from the watchlist.
3. `GET /currency/price/{coin_symbol}` - get most recent price
  - `GET /currency/price/{coin_symbol}?timestamp=...` - get price closest to the specified timestamp

In my opinion, REST endpoints would be a better fit:
- POST `/currency_rest/{name}`
- DELETE `/currency_rest/{name}`
- GET `/currency_rest/{name}[?timestamp={timestamp}]`

golangci-lint was used to lint the source code.

# TODO
- [ ] OpenAPI
  - [ ] OpenAPI 3.0 yaml spec
  - [ ] Use [redocli-cli](https://redocly.com/redocly-cli) in a pre-commit hook (or in a CI pipeline) to validate the spec
  - [ ] Write tests to check if API matches
- [ ] clean up error handling

# Running
Create a `.env` file with the following structure:
```
DB_PASSWORD=example_password
DB_DATABASE=interview
DB_USERNAME=postgres
COIN_GECKO_TOKEN=your-token
POLLING_INTERVAL=5
```

Build the image:
```bash
docker build -t interview-test:latest .
```

Start docker compose:
```bash
docker compose up
```

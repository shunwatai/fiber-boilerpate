# Fiber boilerpate
This project is for learning purpose, just a starter project for myself to learn go and start a REST API quicker.

It is running by fiber with basic CRUD routes which follows the Controller-Service-Repository pattern like Spring or Laravel's structure.

# Features
- With implementations of `postgres`, `sqlite`, `mariadb`, `mongodb` for storing records in DB. Just raw sql without ORM.
- With example modules like `users`, `todos`, `documents` etc. in `interal/modules/`, with CRUD APIs.
- With a [script](#generate-new-module) `cmd/gen/gen.go` for generate new module to `internal/modules/`.
- With the example of JWT auth in the [login API](#login).
- Can generate swagger doc.
- Make use of `viper` for loading env variables in config.
- With a logging wrapper by `zap` which uses as middleware for writing the request's logs in `log/`, the log file may be used for centralised log server like ELK or Signoz. 

# Todo
- [ ] Need more test cases
- [ ] Try `bubbletea` for `cmd/gen/gen.go`
- [ ] Try Oauth
- [ ] Web template by htmx
    - [x] Login page
    - [x] Forget page
    - [x] Users page CRUD
        - [x] Users list
        - [x] User form
    - [ ] Todos page
        - [ ] Todos list
        - [ ] Todo form
        - [ ] Todo form upload file
- [ ] DB try views and joining in repository

# Quick start by docker-compose
1. [Start the databases](#start-databases-for-development)
2. [Run database migrations](#run-migration)
3. [Set the database in config](#for-run-by-docker)
3. [Start fiber api](#start-by-docker)
4. [Test the apis by curl](#test-sample-APIs)

# Install dependencies
If run the Fiber server without docker, install the following go packages.
## Air - hot reload
`air` for fiber hotreload.
```
go install github.com/cosmtrek/air@latest
```

## Go-migrate - db migration
`migrate` for run the database's migration.
```
go install -tags 'postgres mysql sqlite3 mongodb' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Swag - swagger doc
`swag` command for generate swagger doc.
```
go install github.com/swaggo/swag/cmd/swag@latest
```

## Nodejs - tailwindcss CLI build
`npx` for running the tailwindcss command for frontend html template dev

# Config
## Edit config
### For run without docker
```
cp configs/localhost.yaml.sample configs/localhost.yaml
```

### For run by docker
```
cp configs/docker.yaml.sample configs/docker.yaml
```

## Set the db driver
At `database` section, edit the `engine`
```
...
database:
  engine: "postgres/sqlite/mariadb/mongodb"
...
```

# Start databases for development
1. copy and then edit the `db.env` if needed
```
cp db.env.sample db.env
```

2. Start all by docker-compose
Postgres, Mariadb & Mongodb will be started
```
docker-compose -f compose-db.yaml up -d
```

# Start the fiber api server
## For development
Set the `env` to `local` in the `configs/<localhost/docker>.yaml`

### Start without docker
```
air
```

### Start by docker
Run the dev container
```
make docker-dev
```

Check status
```
docker-compose -f compose-dev.yaml ps
```

Watch the log
```
make docker-dev-log
```

## For production
Set the `env` to `prod` in the `configs/<localhost/docker>.yaml`

### Start by docker
Run the production container
```
make docker-prod
```

Check status
```
docker-compose -f compose-prod.yaml ps
```

Watch the log
```
make docker-prod-log
```

# DB Migration
## Create new migration
```
migrate create -ext sql -dir migrations/<dbEngine(postgres/mariadb/sqlite)> -seq <migrationName>
```

e.g.
postgres:
```
migrate create -ext sql -dir migrations/postgres -seq add_new_col_to_users
```
mongodb:
```
migrate create -ext json -dir migrations/mongodb -seq add_xxx_index_to_users
```

## Run migration
### Sqlite
#### Run migrations
```
go run main.go migrate-up sqlite
```
or
```
migrate -source file://migrations/sqlite -database "sqlite3://fiber-starter.db?_auth&_auth_user=user&_auth_pass=user&_auth_crypt=sha1" up
```

#### Revert migration
```
go run main.go migrate-down sqlite
```
or
```
migrate -source file://migrations/sqlite -database "sqlite3://fiber-starter.db?_auth&_auth_user=user&_auth_pass=user&_auth_crypt=sha1" down 1
```

### Mariadb
#### Run migrations
```
go run main.go migrate-up mariadb
```
or
```
migrate -source file://migrations/mariadb -database "mysql://user:password@tcp(localhost:3306)/fiber-starter" up
```

#### Revert migration
```
go run main.go migrate-down mariadb
```
or
```
migrate -source file://migrations/mariadb -database "mysql://user:password@tcp(localhost:3306)/fiber-starter" down 1
```

### Postgres
#### Run migrations
```
go run main.go migrate-up postgres
```
or
```
migrate -source file://migrations/postgres -database "postgres://user:password@localhost:5432/fiber-starter?sslmode=disable" up
```

#### Revert migration
```
go run main.go migrate-down postgres
```
or
```
migrate -source file://migrations/postgres -database "postgres://user:password@localhost:5432/fiber-starter?sslmode=disable" down 1
```

### Mongodb
#### Run migrations
```
go run main.go migrate-up mongodb
```
or
```
migrate -source file://migrations/mongodb -database "mongodb://user:password@localhost:27017/fiber-starter?authSource=admin" up
```

#### Revert migration
```
go run main.go migrate-down mongodb
```
or
```
migrate -source file://migrations/mongodb -database "mongodb://user:password@localhost:27017/fiber-starter?authSource=admin" down 1
```

# Run the tailwindcss build process
```
make tw-watch
```

# Test sample APIs
## ping
```
curl --request GET \
  --url http://localhost:7000/ping
```

## login
```
curl --request POST \
  --url http://localhost:7000/api/auth/login \
  --header 'Content-Type: application/json' \
  --data '{"name":"admin","password":"admin"}'
```

# Generate new module
The `cmd/gen/gen.go` is for generating new module without tedious copy & paste, find & replace.

[Detail usage](cmd/gen/README.md)

# API details
## Password reset
[readme](internal/modules/passwordReset/README.md)

# Run tests
To disable cache when running tests, run with options: `-count=1`
ref: https://stackoverflow.com/a/49999321

## Run all tests
```
go test -v -race ./... -count=1
```

## Run specific database tests

### Run sqlite's tests
```
go test -v ./internal/database -run TestSqliteConstructSelectStmtFromQuerystring -count=1
```

### Run mariadb's tests
```
go test -v ./internal/database -run TestMariadbConstructSelectStmtFromQuerystring -count=1
```

### Run postgres's tests
```
go test -v ./internal/database -run TestPgConstructSelectStmtFromQuerystring -count=1
```

### Run mongodb's tests
```
go test -v ./internal/database -run TestMongodbConstructSelectStmtFromQuerystring -count=1
```

# Swagger
## Edit the doc
In each module under `internal/modules/<module>/route.go`, edit the swagger doc before generate the `docs/` directory at next section below.

## Format swagger's comments & generate the swagger docs
```
$ swag fmt
$ swag init
```

## go to the swagger page by web browser
http://localhost:7000/swagger/index.html

# Send log to Signoz
## Spin up the otel container
```
docker run -d --name signoz-host-otel-collector --user root -v $(pwd)/log/requests.log:/tmp/requests.log:ro -v $(pwd)/otel-collector-config.yaml:/etc/otel/config.yaml --add-host host.docker.internal:host-gateway signoz/signoz-otel-collector:0.88.11
```

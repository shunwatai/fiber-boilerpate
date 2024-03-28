# Fiber boilerpate
It is my toy project for learning go, just a starter project for myself to build REST API.
It is a REST API running by go fiber with basic CRUD which follows the Controller-Service-Repository like Spring or Laravel's structure.

# Features
- With implementations of `postgres`, `sqlite`, `mariadb`, `mongodb` for storing records without ORM.
- With example modules like `users`, `todos`, `documents` etc. in `interal/modules/`, with CRUD APIs.
- With a [simple script](#generate-new-module) `cmd/gen/gen.go` for generate new module to `internal/modules/`.
- With the example of JWT auth in the (login API)[#login].
- Can generate swagger doc.
- With a logging wrapper by `zap` middleware which output the request's logs in `log/`, the log file may be used for centralised log server like ELK or Signoz. 

# Quick start by docker-compose
1. [Start the databases](#start-databases-for-development)
2. [Run database migrations](#run-migration)
3. [Set the database in config](#for-run-by-docker)
3. [Start fiber api](#start-by-docker)
4. [Test the apis by curl](#test-sample-APIs)

# Install dependencies
If run the Fiber server without docker, install the following go packages.
## Air - hot reload
```
go install github.com/cosmtrek/air@latest
```

## Go-migrate - db migration
```
go install -tags 'postgres mysql sqlite3 mongodb' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Swag - swagger doc
```
go install github.com/swaggo/swag/cmd/swag@latest
```

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

# Start the api server
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

## Usage
Module name should be a singular noun, with an initial which uses as the reciver methods.
```
go run main.go generate <module-name-in-singular-lower-case e.g: userDocument> <initial e.g: u (for ud)>
```

Example to generate new module `post`
```
go run main.go generate post p
```
sample output:
```
...
created internal/modules/post

created /home/drachen/git/personal/fiber-starter/migrations/postgres/000009_create_posts.up.sql
created /home/drachen/git/personal/fiber-starter/migrations/postgres/000009_create_posts.down.sql
...
created /home/drachen/git/personal/fiber-starter/migrations/mongodb/000008_create_posts.up.json
created /home/drachen/git/personal/fiber-starter/migrations/mongodb/000008_create_posts.down.json

DB migration files for post created in ./migrations, 
please go to add the SQL statements in up+down files, and then run: make migrate-up
```

Afterwards, the following should be created:
- `interal/module/posts/`
- `migrations/<postgres&mariadb&sqlite&mongodb>/xxxxx_create_posts.<sql/json>`

Then you have to edit the `interal/modules/post/type.go` for its fields,
and edit the migration files in `migrations/<postgres/mariadb/sqlite/mongodb>` for its columns and run the migrations.
Then the `post`'s CRUD should be ready.

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

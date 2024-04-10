# migrations
All the following commands run in project root folder.

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

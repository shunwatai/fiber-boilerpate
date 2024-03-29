# Gen
This package is a go script for generate new module.

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


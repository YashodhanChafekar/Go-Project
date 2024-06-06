# Go-Project

## 1. Create a new project

You should know how to do this by now! My process is:

* Create a repo on GitHub or GitLab (initialized with a README)
* Clone it onto your machine
* Create a new Go module with `go mod init`.
* Create a `main.go` file in the root of your project, and add a `func main()` to it.

## 2. Install packages

Install the following packages using `go get`:

* [chi](https://github.com/go-chi/chi)
* [cors]https://github.com/go-chi/cors
* [godotenv](github.com/joho/godotenv)

## 3. Env

Create a gitignore'd `.env` file in the root of your project and add the following:

```bash
PORT="8080"
```

The `.env` file is a convenient way to store environment (configuration) variables.

* Use [godotenv.Load()](https://pkg.go.dev/github.com/joho/godotenv#Load) to load the variables from the file into your environment at the top of `main()`.
* Use [os.Getenv()](https://pkg.go.dev/os#Getenv) to get the value of `PORT`.

## 4. Create a router and server

1. Create a [chi.NewRouter](https://pkg.go.dev/github.com/go-chi/chi#NewRouter)
2. Use [router.Use](https://pkg.go.dev/github.com/go-chi/chi#Router.Use) to add the built-in [cors.Handler](https://pkg.go.dev/github.com/go-chi/cors#Handler) middleware.
3. Create sub-router for the `/v1` namespace and mount it to the main router.
4. Create a new [http.Server](https://pkg.go.dev/net/http#Server) and add the port and the main router to it.
5. Start the server

## 5. Create some JSON helper functions

Create two functions:

* `respondWithJSON(w http.ResponseWriter, status int, payload interface{})`
* `respondWithError(w http.ResponseWriter, code int, msg string)` (which calls `respondWithJSON` with error-specific values)

You used these in the "Learn Web Servers" course, so you should be able to figure out how to implement them again. They're simply helper functions that write an HTTP response with:

* A status code
* An `application/json` content type
* A JSON body

## 6. Add a readiness handler

Add a handler for `GET /v1/healthz` requests. It should return a 200 status code and a JSON body:

```json
{
  "status": "ok"
}
```

*The purpose of this endpoint is for you to test your `respondWithJSON` function.*

## 7. Add an error handler

Add a handler for `GET /v1/err` requests. It should return a 500 status code and a JSON body:

```json
{
  "error": "Internal Server Error"
}
```

*The purpose of this endpoint is for you to test your `respondWithError` function.*

## 8. Run and test your server

```bash
go build -o out && ./out
```

Once it's running, use an HTTP client to test your endpoints.

# Create Users

In this step, we'll be adding an endpoint to create new users on the server. We'll be using a couple of tools to help us out:

* [database/sql](https://pkg.go.dev/database/sql): This is part of Go's standard library. It provides a way to connect to a SQL database, execute queries, and scan the results into Go types.
* [sqlc](https://sqlc.dev/): SQLC is an *amazing* Go program that generates Go code from SQL queries. It's not exactly an [ORM](https://www.freecodecamp.org/news/what-is-an-orm-the-meaning-of-object-relational-mapping-database-tools/), but rather a tool that makes working with raw SQL almost as easy as using an ORM.
* [Goose](https://github.com/pressly/goose): Goose is a database migration tool written in Go. It runs migrations from the same SQL files that SQLC uses, making the pair of tools a perfect fit.

## 1. Install SQLC

SQLC is just a command line tool, it's not a package that we need to import. I recommend [installing](https://docs.sqlc.dev/en/latest/overview/install.html) it using `go install`:

```bash
go install github.com/kyleconroy/sqlc/cmd/sqlc@latest
```

Then run `sqlc version` to make sure it's installed correctly.

## 2. Install Goose

Like SQLC, Goose is just a command line tool. I also recommend [installing](https://github.com/pressly/goose#install) it using `go install`:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Run `goose -version` to make sure it's installed correctly.

## 3. Create the `users` migration

Create an `sql` directory in the root of your project, and in there creating a `schema` directory.

A "migration" is a SQL file that describes a change to your database schema. For now, we need our first migration to create a `users` table. The simplest format for these files is:

```
number_name.sql
```

For example, I created a file in `sql/schema` called `001_users.sql` with the following contents:

```sql
-- +goose Up
CREATE TABLE ...

-- +goose Down
DROP TABLE users;
```

Write out the `CREATE TABLE` statement in full, I left it blank for you to fill in. A `user` should have 4 fields:

* id: a `UUID` that will serve as the primary key
* created_at: a `TIMESTAMP` that can not be null
* updated_at: a `TIMESTAMP` that can not be null
* name: a string that can not be null

The `-- +goose Up` and `-- +goose Down` comments are required. They tell Goose how to run the migration. An "up" migration moves your database from its old state to a new state. A "down" migration moves your database from its new state back to its old state.

By running all of the "up" migrations on a blank database, you should end up with a database in a ready-to-use state. "Down" migrations are only used when you need to roll back a migration, or if you need to reset a local testing database to a known state.

## 4. Run the migration

`cd` into the `sql/schema` directory and run:

```bash
goose postgres CONN up
```

Where `CONN` is the connection string for your database. Here is mine:

```
postgres://yashodhan:yashodhan@localhost:5432/blogator
```

The format is:

```
protocol://username:password@host:port/database
```

Run your migration! Make sure it works by using PGAdmin to find your newly created `users` table.

## 5. Save your connection string as an environment variable

Add your connection string to your `.env` file. When using it with `goose`, you'll use it in the format we just used. However, here in the `.env` file it needs an additional query string:

```
protocol://username:password@host:port/database?sslmode=disable
```

Your application code needs to know to not try to use SSL locally.

## 6. Configure [SQLC](https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html)

Always run the `sqlc` command from the root of your project. Create a file called `sqlc.yaml` in the root of your project. Here is mine:

```yaml
version: "2"
sql:
  - schema: "sql/schema"
    queries: "sql/queries"
    engine: "postgresql"
    gen:
      go:
        out: "internal/database"
```

We're telling SQLC to look in the `sql/schema` directory for our schema structure (which is the same set of files that Goose uses, but sqlc automatically ignores "down" migrations), and in the `sql/queries` directory for queries. We're also telling it to generate Go code in the `internal/database` directory.

## 7. Write a query to create a user

Inside the `sql/queries` directory, create a file called `users.sql`. Here is mine:

```sql
-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES ($1, $2, $3, $4)
RETURNING *;
```

`$1`, `$2`, `$3`, and `$4` are parameters that we'll be able to pass into the query in our Go code. The `:one` at the end of the query name tells SQLC that we expect to get back a single row (the created user).

Keep the [SQLC docs](https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html) handy, you'll probably need to refer to them again later.

## 8. Generate the Go code

Run `sqlc generate` from the root of your project. It should create a new package of go code in `internal/database`.

## 9. Open a connection to the database, and store it in a config struct

it's common to use a "config" struct to store shared data that HTTP handlers need access to. We'll do the same thing here. Mine looks like this:

```go
type apiConfig struct {
	DB *database.Queries
}
```

At the top of `main()` load in your database URL from your `.env` file, and then [.Open()](https://pkg.go.dev/database/sql#Open) a connection to your database:

```go
db, err := sql.Open("postgres", dbURL)
```

Use your generated `database` package to create a new `*database.Queries`, and store it in your config struct:

```go
dbQueries := database.New(db)
```

## 10. Create an HTTP handler to create a user

Endpoint: `POST /v1/users`

Example body:

```json
{
  "name": "Yashodhan"
}
```

Example response:

```json
{
    "id": "3f8805e3-634c-49dd-a347-ab36479f3f83",
    "created_at": "2021-09-01T00:00:00Z",
    "updated_at": "2021-09-01T00:00:00Z",
    "name": "Yashodhan"
}
```

Use Google's [UUID](https://pkg.go.dev/github.com/google/uuid) package to generate a new [UUID](https://blog.boot.dev/clean-code/what-are-uuids-and-should-you-use-them/) for the user's ID. Both `created_at` and `updated_at` should be set to the current time. If we ever need to update a user, we'll update the `updated_at` field.

# API Key

## 1. Add an "api key" column to the users table

Use a new migration file in the `sql/schema` directory to add a new column to the `users` table. I named my file `002_users_apikey.sql`.

The "up" migration adds the column, and the "down" migration removes it.

Use a `VARCHAR(64)` that must be unique and not null. Using a string of a specific length does two things:

1. It ensures we don't accidentally store a key that's too long (type safety)
2. It's more performant than using a variable length `TEXT` column

Because we're enforcing the `NOT NULL` constraint, and we already have some users in the database, we need to set a default value for the column. A blank default would be a bit silly: that's no better than null! Instead, we'll generate valid API keys (256-bit hex values) using SQL. Here's the function I used:

```sql
encode(sha256(random()::text::bytea), 'hex')
```

When you're done, use `goose postgres CONN up` to perform the migration.

## 2. Create an API key for new users

Update your "create user" SQL query to use the same SQL function to generate API keys for new users.

## 3. Add a new SQL query to get a user by their API key

This query can live in the same file as the "create user" query, or you can make a new one - it's up to you.

## 4. Generate new Go code

Run `sqlc generate` to generate new Go code for your queries.

## 5. New endpoint: return the current user

Add a new endpoint that allows users to get their own user information. You'll need to parse the header and use your new query to get the user data.

Endpoint: `GET /v1/users`

Request headers: `Authorization: ApiKey <key>`

Example response body:

```json
{
    "id": "3f8805e3-634c-49dd-a347-ab36479f3f83",
    "created_at": "2021-09-01T00:00:00Z",
    "updated_at": "2021-09-01T00:00:00Z",
    "name": "Lane",
    "api_key": "cca9688383ceaa25bd605575ac9700da94422aa397ef87e765c8df4438bc9942"
}
```

# Create a Feed

An RSS feed is just a URL that points to some XML. Users will be able to add feeds to our database so that our server (in a future step) can go download all of the posts in the feed (like blog posts or podcast episodes).

## 1. Create a feeds table

Like any table in our DB, we'll need the standard `id`, `created_at`, and `updated_at` fields. We'll also need a few more:

* `name`: The name of the feed (like "The Changelog, or "The Boot.dev Blog")
* `url`: The URL of the feed
* `user_id`: The ID of the user who added this feed

I'd recommend making the `url` field unique so that in the future we aren't downloading duplicate posts. I'd also recommend using [ON DELETE CASCADE](https://stackoverflow.com/a/14141354) on the `user_id` foreign key so that if a user is deleted, all of their feeds are automatically deleted as well.

Write the appropriate migrations and run them.

## 2. Add a query to create a feed

Add a new query to create a feed, then use `sqlc generate` to generate the Go code.

## 3. Create some authentication middleware

Most of the endpoints going forward will require a user to be logged in. Let's DRY up our code by creating some middleware that will check for a valid API key.

### A custom type for handlers that require authentication

```go
type authedHandler func(http.ResponseWriter, *http.Request, database.User)
```

### Middleware that authenticates a request, gets the user and calls the next authed handler

```go
func (cfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
    ///
}
```

### Using the middleware

```go
v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerUsersGet))
```

## 4. Create a handler to create a feed

Create a handler that creates a feed. This handler *and* the "get user" handler should use the authentication middleware.

Endpoint: `POST /v1/feeds`

Example request body:

```json
{
  "name": "The Boot.dev Blog",
  "url": "https://blog.boot.dev/index.xml"
}
```

Example response body:

```json
{
  "id": "4a82b372-b0e2-45e3-956a-b9b83358f86b",
  "created_at": "2021-05-01T00:00:00Z",
  "updated_at": "2021-05-01T00:00:00Z",
  "name": "The Boot.dev Blog",
  "url": "https://blog.boot.dev/index.xml",
  "user_id": "d6962597-f316-4306-a929-fe8c8651671e"
}
```

# Feed Follows

Aside from just adding new feeds to the database, users can specify *which* feeds they want to follow. This will be important later when we want to show users a list of posts from the feeds they follow.

Add support for the following endpoints, and update the "create feed" endpoint as specified below.

## What is a "feed follow"?

A feed follow is just a link between a user and a feed. It's a [many-to-many](https://en.wikipedia.org/wiki/Many-to-many_(data_model)) relationship, so a user can follow many feeds, and a feed can be followed by many users.

Creating a feed follow indicates that a user is now following a feed. Deleting it is the same as "unfollowing" a feed.

It's important to understand that the `ID` of a feed follow is not the same as the `ID` of the feed itself. Each user/feed pair will have a unique feed follow id.

## Create a feed follow

Endpoint: `POST /v1/feed_follows`

*Requires authentication*

Example request body:

```json
{
  "feed_id": "4a82b372-b0e2-45e3-956a-b9b83358f86b"
}
```

Example response body:

```json
{
  "id": "c834c69e-ee26-4c63-a677-a977432f9cfa",
  "feed_id": "4a82b372-b0e2-45e3-956a-b9b83358f86b",
  "user_id": "0e4fecc6-1354-47b8-8336-2077b307b20e",
  "created_at": "2017-01-01T00:00:00Z",
  "updated_at": "2017-01-01T00:00:00Z"
}
```

## Delete a feed follow

Endpoint: `DELETE /v1/feed_follows/{feedFollowID}`

## Get all feed follows for a user

Endpoint: `GET /v1/feed_follows`

*Requires authentication*

Example response:

```json
[
  {
    "id": "c834c69e-ee26-4c63-a677-a977432f9cfa",
    "feed_id": "4a82b372-b0e2-45e3-956a-b9b83358f86b",
    "user_id": "0e4fecc6-1354-47b8-8336-2077b307b20e",
    "created_at": "2017-01-01T00:00:00Z",
    "updated_at": "2017-01-01T00:00:00Z"
  },
  {
    "id": "ad752167-f509-4ff3-8425-7781090b5c8f",
    "feed_id": "f71b842d-9fd1-4bc0-9913-dd96ba33bb15",
    "user_id": "0e4fecc6-1354-47b8-8336-2077b307b20e",
    "created_at": "2017-01-01T00:00:00Z",
    "updated_at": "2017-01-01T00:00:00Z"
  }
]
```

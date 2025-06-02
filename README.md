# Description

This is a project guided by [Boot.dev](https://www.boot.dev).
Users are able to fetch RSS Feeds using the `gator` CLI in the terminal.

## Requirements

- [PostgreSQL](https://www.postgresql.org/)
- [Go](https://go.dev/)

## Install

    go install https://github.com/Corogura/gator@latest

or clone the code and run in the `gator` directory

    go install

## Setup

Create a `.gatorconfig.json` file in the home directory(`~/`).
```
{
    "db_url":"postgres://<username>:<password>@localhost:5432/gator?sslmode=disable",
    "current_user_name":""
}
```
`db_url` is the connection string of the installed PostgreSQL database attached with `?sslmode=disable`.
For Linux users, username is `postgres` and password is set in the next step.

Run `CREATE DATABASE gator;` in PostgreSQL and set the user password `ALTER USER postgres PASSWORD 'postgres';` if using Linux.

## Commands

- `setup` : Run an initial database setup. Run this command once before start using.
- `register`: Creates a new user using the argument as the username. Ex.`register <username>`
- `login` : Log in as a already registered user. (Changes `current_user_name` in the config file to the specified username). Ex.`login <username>`
- `users` : Display a list of all registered users.
- `addfeed` : Adds an RSS Feed using the URL. Ex.`addfeed <feed_name> <feed_url>`
- `feeds` : Display a list of all the feeds in the database.
- `follow` : Follow a feed on the database (potentially created by other users). Ex.`follow <feed_name>`
- `unfollow` : Unfollow a feed. Ex.`unfollow <feed_name>`
- `following` : Display a list of feeds that the current user follows.
- `agg` : Fetch all feeds one by one starting from the most outdated feed, taking a duration as the argument. Ex.`agg 1m0s`
- `browse` : Browse the fetched posts from the feeds that the current user follows with a specified number of posts (default=2). Ex.`browse 3`
- `reset` : Erases all data from the database. Use at caution.
# Gator

You will need Postgres and Go installed to run the program

## Installation

go install github.com/username/gator@latest

## Run

gator

## Setting up config file

In the Home directory create a ".gatorconfig.json"
with a "db_url":""
which will be the postgres database url you will use for your application.

## Gator commands

## Commands and Descriptions

- **`login`**

  - **Description**: Log in with a username.
  - **Arguments**: `<username>`
  - **Example**: `gator login michael`

- **`register`**

  - **Description**: Create a new user account with a username.
  - **Arguments**: `<username>`
  - **Example**: `gator register michael`

- **`reset`**

  - **Description**: Reset the user database.
  - **Arguments**: None
  - **Example**: `gator reset`

- **`users`**

  - **Description**: List all users and show the currently logged-in user.
  - **Arguments**: None
  - **Example**: `gator users`

- **`agg`**

  - **Description**: Fetch aggregated content from subscribed feeds after a specified time.
  - **Arguments**: `<time>` (e.g., `1m` for 1 minute)
  - **Example**: `gator agg 1m`

- **`addfeed`**

  - **Description**: Add a new feed to the user's subscriptions.
  - **Arguments**: `<name> <url>`
  - **Example**: `gator addfeed TechCrunch https://techcrunch.com/rss`

- **`feeds`**

  - **Description**: Display all available feeds.
  - **Arguments**: None
  - **Example**: `gator feeds`

- **`follow`**

  - **Description**: Follow a user's feed that is already in the database.
  - **Arguments**: `<feed-url>`
  - **Example**: `gator follow https://example.com/feed`

- **`unfollow`**

  - **Description**: Unfollow a user's feed.
  - **Arguments**: `<feed-url>`
  - **Example**: `gator unfollow https://example.com/feed`

- **`browse`**
  - **Description**: Browse content from feeds with a limit on the number of items.
  - **Arguments**: `<limit>`
  - **Example**: `gator browse 10`

# GH Estimate Bot

A tiny GitHub App that comments on newly opened issues if they lack an `Estimate: X days` field.

## Set Up GitHub App

Create a Github App: Settings → Developer settings → GitHub Apps → New GitHub App

In the New Github App form, you will be asked for a Webhook secret, which you can generate like this:

```bash
openssl rand -hex 30
```

When prompted for a webhook endpoint, you can use a seervice like [Smee](https://smee.io/), [Pinggy](https://pinggy.io/), [Ngrok](https://ngrok.com/), etc.

Your all will need `Issues: Read & write` permission, and it also needs to subscribe to `Issues` events.
After you submit your GH App creation form, you will need to generate a private key (a `.pem` file).

Once all that is done, create your .env file:

```bash
cp .env.example .env
```

Then, make sure to update it with your GH App ID, Webhook secret, and the location of the `.pem` file.

## Run locally

Run the server via:

```bash
make run
```

This will run your server on port 8080 (or whatever port you specified in .env).

The server exposes 2 endpoints:

- `GET /_/health`: Responds with `HTTP 200 (OK)` if the server is running
- `POST /wh`: Handles incoming events from GitHub

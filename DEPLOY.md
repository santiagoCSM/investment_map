# Deploying to Railway

This app deploys as a single Docker image: the Go binary serves both the
REST API and the built React frontend (see the root `Dockerfile`). It needs
two external services in production — Redis and MongoDB — since Railway
doesn't run persistent databases for you on the free tier.

Everything below after "1. Create the databases" happens in your browser,
in accounts only you can create — sign-ups, payment-free tier selection,
and connecting your GitHub account aren't things that can be scripted from
outside.

## 1. Create the databases

### Redis — Upstash

1. Go to [upstash.com](https://upstash.com) and sign up (GitHub login is fine).
2. Create a new **Redis** database. Pick a region close to wherever you'll
   deploy on Railway (doesn't have to be exact — just close).
3. On the database's page, find the **Connect** / **REST API** section and
   copy the **`rediss://` connection string** (note the double "s" — that's
   the TLS variant, which is what this app expects). It looks like:
   ```
   rediss://default:AbCdEf123...@usw1-abc-12345.upstash.io:6379
   ```
4. Save that as `REDIS_URL` for step 3.

Upstash's free tier is enough for this app — it's a handful of keys
refreshed every few minutes, nowhere near their request/storage limits.

### MongoDB — Atlas

1. Go to [mongodb.com/cloud/atlas](https://www.mongodb.com/cloud/atlas) and
   sign up.
2. Create a free **M0** cluster (no credit card required for M0).
3. Under **Database Access**, create a database user (username + password —
   generate a strong password, you'll paste it into the connection string).
4. Under **Network Access**, add `0.0.0.0/0` (allow access from anywhere).
   Railway's outbound IPs aren't fixed on the free plan, so a specific
   allowlist isn't practical here — this is the standard tradeoff for
   PaaS-to-Atlas connections without a static IP or VPC peering.
5. Click **Connect** on your cluster → **Drivers** → copy the connection
   string. It looks like:
   ```
   mongodb+srv://<username>:<password>@cluster0.abcde.mongodb.net/?retryWrites=true&w=majority
   ```
6. Replace `<username>`/`<password>` with the real values from step 3.
   Save that as `MONGO_URI` for step 3.

## 2. Push this repo to GitHub

If you haven't already, make sure the latest commit (with the `Dockerfile`,
Redis URL support, and static-file serving) is pushed to
`github.com/nobodywhois0/investment_map`.

## 3. Create the Railway service

1. Go to [railway.app](https://railway.app) and sign up/log in with GitHub.
2. **New Project** → **Deploy from GitHub repo** → select
   `nobodywhois0/investment_map`. Railway will detect the root `Dockerfile`
   automatically and build from it — no extra config needed there.
3. Once the service is created, open its **Variables** tab and add:

   | Variable | Value |
   |---|---|
   | `REDIS_URL` | the `rediss://...` string from Upstash |
   | `MONGO_URI` | the `mongodb+srv://...` string from Atlas |
   | `MONGO_DATABASE` | `investment_map` |
   | `REFRESH_INTERVAL` | `5m` (optional — this is already the default) |
   | `HISTORY_REFRESH_INTERVAL` | `24h` (optional — already the default) |

   Don't set `PORT` or `HTTP_ADDR` — Railway injects `PORT` itself, and the
   app already reads it automatically.

4. Under **Settings → Networking**, click **Generate Domain** to get a
   public `*.up.railway.app` URL.
5. Deploy (Railway does this automatically on push once connected). Watch
   the **Deploy Logs** — you should see the same `worker starting` /
   `http server listening` lines you see locally with `make run`.

## 4. Verify it's actually working

Once deployed, from your own machine:

```bash
curl https://<your-app>.up.railway.app/healthz
curl https://<your-app>.up.railway.app/api/quotes
```

The first request after deploy may return fewer than 25 quotes — the
workers need one full tick (up to `REFRESH_INTERVAL`, 5 min by default) to
populate every symbol, and the very first run will also trigger the Mongo
history backfill (25 symbols × ~2500 points each), which takes a couple of
minutes. Everything after that first tick should be fully populated. Then
open `https://<your-app>.up.railway.app/` in a browser — you should see the
full map.

## Notes on cost

- Upstash and Atlas M0 are both genuinely free at this app's scale (a
  couple dozen keys/documents, refreshed every few minutes).
- Railway's free tier is a monthly usage credit, not unlimited — a
  small always-on service like this one will consume some of that credit
  continuously. Check Railway's current pricing page for the exact numbers
  before leaving it running long-term; this project doesn't need to run
  24/7 to be useful as a portfolio piece, so pausing the service between
  demos is a reasonable way to stay inside the free credit.

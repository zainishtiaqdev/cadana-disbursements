# Deploy — live demo

Three free pieces: **Supabase** (Postgres) · **Render** (Go backend) · **Vercel** (Vue frontend).
No VPS, no Docker, no paid domain — each platform hands you a free subdomain.

```
Vercel (SPA)  ──HTTPS──►  Render (Go API)  ──TLS──►  Supabase (Postgres)
your.vercel.app           your.onrender.com          db / pooler .supabase.co
```

> ⚠️ **Use the Supabase Session pooler URI for Render — not the direct host.**
> The direct host (`db.<ref>.supabase.co`) is **IPv6-only** on the free tier; Render's egress can't
> reach it reliably. The **Session pooler** (`...pooler.supabase.com`, IPv4) works everywhere.
> Supabase → **Connect** → **Connection string** → **URI** → **Session pooler**.

---

## 1 · Supabase (already created)

Project ref: `vxkvucfsbffgqlyzcprx`. The app creates its own tables on first boot — nothing to run by
hand. Grab the **Session pooler** URI; it looks like:

```
postgresql://postgres.vxkvucfsbffgqlyzcprx:[YOUR-DB-PASSWORD]@aws-0-<region>.pooler.supabase.com:5432/postgres
```

(Reset the DB password under **Settings → Database** if you don't have it. Keep it out of git.)

## 2 · Backend → Render

**Option A — Blueprint:** the repo ships [`render.yaml`](./render.yaml). In Render: **New → Blueprint**,
pick this repo, then set the `DATABASE_URL` secret when prompted.

**Option B — manual Web Service:**
- Root directory: `backend`
- Build: `go build -o server ./cmd/server`
- Start: `./server`
- Health check path: `/healthz`
- Env vars:
  - `DATABASE_URL` = the Session pooler URI from step 1 (append `?sslmode=require`)
  - `ALLOWED_ORIGIN` = `*` (or your Vercel URL once you have it)
  - `PORT` is injected by Render automatically.
- Plan: **Free**.

> Free instances sleep after ~15 min idle and cold-start in ~30–60s. State lives in Supabase, so
> **nothing is lost** — the first request after idle is just slow. Want always-on? Use Fly.io (needs a
> card) with the same env vars.

Note the backend URL, e.g. `https://cadana-disbursements-api.onrender.com`.

## 3 · Frontend → Vercel

- **New Project** → import the repo.
- Root directory: `frontend` · Framework preset: **Vite** (auto-detected).
- Build: `npm run build` · Output: `dist` (both auto-detected).
- Env var: `VITE_API_BASE` = your Render backend URL (no trailing slash).
- Deploy → you get `https://<project>.vercel.app`.

## 4 · Tighten CORS (optional)

Set the backend's `ALLOWED_ORIGIN` to the Vercel URL and redeploy, so the API only answers your
frontend. `*` is fine for a throwaway demo.

## 5 · Smoke test

```bash
curl https://<your-backend>.onrender.com/workers          # 9 workers
# then open https://<project>.vercel.app and run a batch
```

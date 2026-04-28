<h1 align="center">
  <br>
  <a href="https://corecheck.dev"><img src="https://github.com/bitcoin-coverage/core/raw/master/docs/assets/logo.png" alt="Bitcoin Coverage" width="200"></a>
  <br>
    Bitcoin Coverage's front-end
  <br>
</h1>

<h4 align="center">Bitcoin Coverage's front-end built with <a href="https://svelte.dev" target="_blank">Svelte</a>.</h4>

## 📖 Introduction
This repository contains the front-end of Bitcoin Coverage. It is built with [Svelte](https://svelte.dev) and [SvelteKit](https://kit.svelte.dev).

## 🚀 Developement

```bash
npm install # or pnpm install or yarn
npm run dev

# or start the server and open the app in a new browser tab
npm run dev -- --open
```

## 🔧 Environment

Copy `.env.example` to `.env` before local development or builds:

```bash
cp .env.example .env
```

- `PUBLIC_ENDPOINT`: CoreCheck API base URL used by the app.
- `PUBLIC_DASHBOARD_GITHUB_URL`
- `PUBLIC_DASHBOARD_TESTS_URL`
- `PUBLIC_DASHBOARD_BENCHMARKS_URL`
- `PUBLIC_DASHBOARD_JOBS_URL`

The `PUBLIC_DASHBOARD_*` variables override the public Grafana dashboard URLs
used by the app. If left blank, the frontend falls back to the deterministic
Grafana hostnames provisioned by Terraform (`grafana.corecheck.dev` for default
and `grafana-dev.corecheck.dev` for dev). The Terraform stack also exposes a
`public_dashboard_env_overrides` output with the exact deployed URLs if you want
to inject them explicitly at runtime.

## 📦 Build

```bash
npm run build
```

## 📝 License

MIT - [Aurèle Oulès](https://github.com/aureleoules)

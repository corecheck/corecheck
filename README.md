# Corecheck

**Live site: [corecheck.dev](https://corecheck.dev)**

Corecheck is the repository behind **corecheck.dev**, a web app for following Bitcoin Core pull requests through coverage, benchmarks, mutation testing, and job status. It pulls the useful bits into one place so you can see what changed without bouncing between CI logs, GitHub, and one-off reports.

## Using Corecheck

Start on **[corecheck.dev](https://corecheck.dev)** and open the pull request view for the latest Bitcoin Core work. From there you can:

- browse tracked pull requests
- open a report for a PR and compare it against the recorded master baseline
- review benchmark changes tied to the same base commit
- inspect mutation testing results down to the affected file and line
- check the dedicated pages for tests, benchmarks, master coverage, and job activity

## Main features

- PR coverage reports with new-code and baseline coverage breakdowns
- benchmark comparisons anchored to the same base master commit
- mutation testing explorer with per-file drilldown
- master coverage tracking
- pull request and job status views in one place

## How the repo is organized

- `frontend/` - SvelteKit app for the website
- `functions/api/` - Go Lambda APIs for pulls, reports, mutation results, and master coverage
- `functions/compute/` - background jobs that sync GitHub state and kick off analysis work
- `workers/` - batch workers for coverage, benchmarks, mutation testing, and sonar
- `internal/` - shared Go packages for config, database access, logging, and API wiring
- `deploy/terraform/` - infrastructure definitions used to run Corecheck on AWS

## High-level architecture

GitHub data is synced into Corecheck's database, scheduled compute jobs queue analysis work, workers generate the reports and artifacts, and the API layer serves the results to the frontend. Corecheck keeps PR results tied to a known master snapshot, which makes coverage and benchmark comparisons much easier to read.

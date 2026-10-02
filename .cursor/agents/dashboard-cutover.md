---
name: dashboard-cutover
description: Splits the Datadog-to-Grafana migration into small master PRs and keeps the local branch mergeable afterward so the next PR is a smaller diff. Use proactively when cutting a PR from replace-datadog, landing CloudWatch writers, shipping one Grafana dashboard, or rebasing that branch after master moves.
---

You cut small master PRs out of the Datadog-to-Grafana work, then keep the local branch compatible with those PRs so the next slice is still easy to cut.

When invoked:

1. Do not commit, push, or reset unless the user explicitly asks.
2. Compare `master..HEAD` on the current branch. Classify each commit as data collection, dashboard presentation, or mixed. Mixed means one commit both writes CloudWatch and changes Grafana JSON, the Datadog proxy, or a frontend iframe.
3. Do not recommend replaying or cherry-picking the branch as it stands. The first commit is a full cutover: it removes the Datadog proxy, points the frontend iframes at Grafana, and switches writers in one change. Later commits assume that tree.
4. Treat the final tree as the thing to slice. A data-collection change for master must keep the Datadog proxy, the frontend iframes, and the existing Datadog emitters until that specific page is switched. The current writers are CloudWatch-only, so shipping them unchanged turns the live Datadog dashboards stale.
5. Judge each dashboard separately. The site has four routes (`/`, `/tests`, `/benchmarks`, `/jobs`). Grafana can be deployed before any route changes. A route should switch only when its data source already has the history that dashboard needs.
6. After a slice is identified, keep the unpublished branch compatible with it. The local branch should still contain every dashboard and writer that has not shipped, and it should apply cleanly on top of master once that slice is on master. Prefer a rebase onto the updated master over a second copy of the same change. Do not rebase, reset, or force-push unless the user explicitly asks. Until they ask, report the commits that would move and the conflicts you expect.
7. A master PR must not delete or rewrite files the local branch still needs for the remaining dashboards. If a file is shared, the master change is the compatible subset and the branch keeps the rest. After master has the slice, the branch's diff against master should shrink to the unshipped work.

Data sources:

- GitHub overview needs the stats lambda and `/corecheck/github-events/<workspace>`. The first run only writes events from the last 24 hours (`cwLogsMaxAge`). History builds after that. It does not import years of GitHub events.
- Tests need the coverage worker to write `/corecheck/test-results/<workspace>`. Points exist only for runs after that writer is deployed.
- Benchmarks need `handle-benchmarks` to write `/corecheck/benchmark-results/<workspace>`. Same limit.
- Jobs reads existing `AWS/States` and `AWS/Batch` metrics. It does not need a new writer.

Retention on the branch is 731 days for GitHub events and 1096 days for test and benchmark results.

Output:

- Whether the request is feasible, and by slicing the tree rather than cherry-picking
- What must stay on Datadog until cutover so the live site does not go dark
- Which writers can start early, and what history they will not have
- Which dashboard can go live first, and which frontend file switches it
- Commits or files that are mixed and have to be split by hand
- How the local branch stays compatible after the slice: what remains unshipped, and whether a rebase onto master would be clean

# Local Grafana

Runs the public dashboards on your machine, querying the same CloudWatch log groups and metrics as the dev Grafana instance. A dashboard edit no longer replaces the EC2 instance.

```sh
python3 tools/local-grafana/up.py
```

That uses the current AWS CLI credentials and the `dev` workspace in `ap-south-1`. A sidecar refreshes the 15-minute access token for as long as `aws login` stays valid (up to 12 hours). When that login ends, run `aws login` again and reload the dashboard; the containers pick it up without a restart. Stop it with Ctrl-C. `python3 tools/local-grafana/up.py down` removes the containers.

Dashboards:

- http://localhost:3000/d/corecheck-github-overview
- http://localhost:3000/d/corecheck-tests
- http://localhost:3000/d/corecheck-benchmarks
- http://localhost:3000/d/corecheck-jobs

Edit the templates in `deploy/terraform/monitoring/grafana-dashboard-templates/`. `up.py` re-renders them, and Grafana reloads the files within a few seconds. The UI is editable too; the next template render overwrites those UI edits, so copy a finished panel back into the template.

The credentials need CloudWatch Logs Insights and metrics read access in the compute region. `generated/` holds rendered JSON and the refreshed credential file. It is gitignored.

The functional and unit test filters are dropdowns. Grafana cannot fill a variable from a Logs Insights query, so the `test-names` container runs that query and Grafana reads the list through the Infinity datasource. The first option is the test whose median duration rose the most across the selected range. Clicking a test name scrolls to that section's duration graph.

Panel screenshots, for checking how a dashboard actually rendered:

```sh
python3 tools/local-grafana/snap.py tests
python3 tools/local-grafana/snap.py tests --panel 202
```

PNGs are written to `tools/local-grafana/generated/snaps/`. This needs the image renderer started with `up.py`.

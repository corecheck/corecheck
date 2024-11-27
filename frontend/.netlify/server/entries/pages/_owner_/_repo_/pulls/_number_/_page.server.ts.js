import { p as public_env } from "../../../../../../chunks/shared-server.js";
import { _ as _fetchReport, a as _fetchSonarCloudIssues } from "../../../../../../chunks/report.js";
async function _fetchPr(fetch, number) {
  return await fetch(`${public_env.PUBLIC_ENDPOINT}/pulls/${number}`).then((res) => res.json()).catch((err) => {
    console.error(err);
  });
}
async function load({ params, fetch }) {
  const pr = await _fetchPr(fetch, params.number);
  let report, sonarcloud;
  try {
    report = await _fetchReport(fetch, public_env.PUBLIC_ENDPOINT, params.number);
    console.log(report);
    sonarcloud = await _fetchSonarCloudIssues(fetch, params.number, report.commit);
  } catch (e) {
    console.error(e);
  }
  return {
    pr,
    report,
    sonarcloud
  };
}
export {
  _fetchPr,
  load
};

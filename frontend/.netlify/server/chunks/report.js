async function _fetchReport(fetch, endpoint, number, id = void 0) {
  const report = await fetch(`${endpoint}/pulls/${number}/report?id=${id || ""}`).then((res) => res.json()).catch((err) => {
    console.error(err);
  });
  return report;
}
async function _fetchSonarCloudIssues(fetch, number, commit) {
  return fetch(`https://sonarcloud.io/api/issues/search?metricKeys=sqale_index&resolved=false&projects=aureleoules_bitcoin&types=CODE_SMELL&branch=${number}-${commit}`).then(async (res) => {
    if (res.status === 200) {
      const data = await res.json();
      const promises = data.issues.map((issue) => {
        return fetch(`https://sonarcloud.io/api/sources/issue_snippets?issueKey=${issue.key}`).then((res2) => res2.json()).then((res2) => {
          const key = Object.keys(res2)[0];
          return {
            ...issue,
            sources: res2[key].sources
          };
        }).catch((err) => {
          console.error(err);
        });
      });
      return {
        ...data,
        issues: await Promise.all(promises)
      };
    }
    return null;
  }).catch((err) => {
    console.error(err);
  });
}
export {
  _fetchReport as _,
  _fetchSonarCloudIssues as a
};

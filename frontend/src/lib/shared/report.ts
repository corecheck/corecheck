export async function _fetchReport(fetch, endpoint, number, id=undefined) {
    const report = await fetch(`${endpoint}/pulls/${number}/report?id=${id || ''}`)
        .then((res) => res.json())
        .catch((err) => {
            console.error(err);
        });

    return report;
}

export async function _fetchSonarCloudIssues(fetch, number, commit) {
    return fetch(`https://sonarcloud.io/api/issues/search?metricKeys=sqale_index&resolved=false&projects=aureleoules_bitcoin&types=CODE_SMELL&branch=${number}-${commit}`)
        .then(async res => {
            if (res.status === 200) {
                const data = await res.json();

                const promises = data.issues.map(issue => {
                    return fetch(`https://sonarcloud.io/api/sources/issue_snippets?issueKey=${issue.key}`)
                        .then(res => res.json())
                        .then(res => {
                            const key = Object.keys(res)[0];
                            return {
                                ...issue,
                                sources: res[key].sources
                            }
                        })
                        .catch(err => {
                            console.error(err);
                        });
                });

                return {
                    ...data,
                    issues: await Promise.all(promises)
                }
            }
            return null;
        })
        .catch((err) => {
            console.error(err);
        });
}
import { env } from '$env/dynamic/public'
import { _fetchReport, _fetchSonarCloudIssues } from '@/lib/shared/report';

export async function _fetchPr(fetch, number) {
    return await fetch(`${env.PUBLIC_ENDPOINT}/pulls/${number}`)
        .then((res) => res.json())
        .catch((err) => {
            console.error(err);
        });
}

export async function load({ params, fetch }) {
    const pr = await _fetchPr(fetch, params.number);

    let report, sonarcloud;
    try {
        report = await _fetchReport(fetch, env.PUBLIC_ENDPOINT, params.number);
        console.log(report);
        sonarcloud = await _fetchSonarCloudIssues(fetch, params.number, report.commit);
    } catch (e) {
        console.error(e);
    }
    return {
        pr,
        report,
        sonarcloud,
    }
}
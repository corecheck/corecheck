import { env } from '$env/dynamic/public';

const publicGrafanaBaseUrl =
    env.PUBLIC_ENDPOINT?.includes('api-dev.')
        ? 'https://grafana-dev.corecheck.dev'
        : 'https://grafana.corecheck.dev';

const fallbackPublicDashboardUrls = {
    github: `${publicGrafanaBaseUrl}/d/corecheck-github-overview/corecheck-github-overview?orgId=1&kiosk`,
    tests: `${publicGrafanaBaseUrl}/d/corecheck-tests/corecheck-tests?orgId=1&kiosk`,
    benchmarks: `${publicGrafanaBaseUrl}/d/corecheck-benchmarks/corecheck-benchmarks?orgId=1&kiosk`,
    jobs: `${publicGrafanaBaseUrl}/d/corecheck-jobs/corecheck-jobs?orgId=1&kiosk`
} as const;

function getDashboardUrl(url: string | undefined, fallbackUrl: string) {
    return url?.trim() || fallbackUrl;
}

export const publicDashboardUrls = {
    github: getDashboardUrl(env.PUBLIC_DASHBOARD_GITHUB_URL, fallbackPublicDashboardUrls.github),
    tests: getDashboardUrl(env.PUBLIC_DASHBOARD_TESTS_URL, fallbackPublicDashboardUrls.tests),
    benchmarks: getDashboardUrl(env.PUBLIC_DASHBOARD_BENCHMARKS_URL, fallbackPublicDashboardUrls.benchmarks),
    jobs: getDashboardUrl(env.PUBLIC_DASHBOARD_JOBS_URL, fallbackPublicDashboardUrls.jobs)
} as const;

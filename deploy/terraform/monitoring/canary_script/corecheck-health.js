const synthetics = require('Synthetics');
const log = require('SyntheticsLogger');

const API_BASE = 'https://api.corecheck.dev';
const FRONTEND_BASE = 'https://corecheck.dev/bitcoin/bitcoin/pulls';
const MAX_CANDIDATES = 5;
const MAX_CODE_UPDATE_AGE_MS = 24 * 60 * 60 * 1000;
const MIN_CODE_UPDATE_AGE_MS = 2 * 60 * 60 * 1000;
const REPORT_CLOCK_SKEW_MS = 5 * 60 * 1000;

function describePR(pr) {
  const title = pr.title ? `: "${pr.title}"` : '';
  return `PR #${pr.number}${title}`;
}

function parseTimestamp(value, label) {
  const timestamp = Date.parse(value);
  if (Number.isNaN(timestamp)) {
    throw new Error(`Invalid ${label} timestamp: ${value}`);
  }

  return timestamp;
}

async function fetchJson(url, errorPrefix) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`${errorPrefix} returned HTTP ${response.status}`);
  }

  return response.json();
}

async function verifyFrontend(page, pr) {
  const prUrl = `${FRONTEND_BASE}/${pr.number}`;
  log.info(`Navigating to ${prUrl}`);

  const navResponse = await page.goto(prUrl, { waitUntil: 'networkidle2', timeout: 30000 });
  if (!navResponse || navResponse.status() >= 400) {
    throw new Error(
      `Frontend returned HTTP ${navResponse ? navResponse.status() : 'null'} for ${prUrl}`,
    );
  }

  await page.waitForSelector('h1', { timeout: 15000 });

  const prTitle = await page.evaluate(() => {
    const h1 = document.querySelector('h1');
    return h1 ? h1.textContent.trim() : null;
  });

  if (!prTitle) {
    throw new Error('PR title heading not found - page may not have loaded correctly');
  }

  await page.waitForSelector('.form-field', { timeout: 15000 });

  const hasCoverageField = await page.evaluate(() => {
    const labels = document.querySelectorAll('.form-field label .txt');
    return Array.from(labels).some((el) => el.textContent.trim() === 'Coverage report');
  });

  if (!hasCoverageField) {
    throw new Error('Coverage report field not found on the PR page');
  }

  await page.waitForSelector('.accordion', { timeout: 20000 });

  const accordionCount = await page.evaluate(() => {
    return document.querySelectorAll('.accordion').length;
  });

  if (accordionCount === 0) {
    throw new Error('No coverage file accordions found - report data may be missing');
  }

  return { prTitle, accordionCount };
}

exports.handler = async function () {
  // Step 1: Fetch the PR list and gather recent candidates whose code changed 2-24 hours ago.
  // The list is sorted by code_updated_at (only set when the HEAD commit changes), so we can
  // try the most recent eligible PRs first and succeed as soon as one has a fresh, renderable report.
  log.info('Fetching PR list from API...');
  const pulls = await fetchJson(`${API_BASE}/pulls?page=1&title=`, 'Pulls API');
  if (!Array.isArray(pulls) || pulls.length === 0) {
    throw new Error('No PRs returned from API - API or database may be down');
  }

  const now = Date.now();
  const newestEligibleUpdate = now - MIN_CODE_UPDATE_AGE_MS;
  const oldestEligibleUpdate = now - MAX_CODE_UPDATE_AGE_MS;

  const candidatePRs = pulls.filter((pr) => {
    if (!pr.code_updated_at) return false;

    const codeUpdatedAt = Date.parse(pr.code_updated_at);
    if (Number.isNaN(codeUpdatedAt)) return false;

    return codeUpdatedAt >= oldestEligibleUpdate && codeUpdatedAt <= newestEligibleUpdate;
  }).slice(0, MAX_CANDIDATES);

  if (candidatePRs.length === 0) {
    throw new Error(
      'No PR found with a code change in the last 24 hours that is at least 2 hours old - ' +
        'the coverage pipeline may be stalled or there has been no recent PR activity',
    );
  }

  log.info(
    `Evaluating ${candidatePRs.length} recent candidate PR(s): ` +
      candidatePRs.map((pr) => `#${pr.number}`).join(', '),
  );

  const candidateFailures = [];
  let page;

  for (const targetPR of candidatePRs) {
    const prLabel = describePR(targetPR);

    try {
      const codeUpdatedAtMs = parseTimestamp(targetPR.code_updated_at, `${prLabel} code_updated_at`);
      log.info(`Checking ${prLabel} (code updated ${targetPR.code_updated_at})`);

      // Step 2: Verify the PR has a fresh successful report tied to this code update.
      const prData = await fetchJson(`${API_BASE}/pulls/${targetPR.number}`, `PR detail API for ${prLabel}`);
      if (!Array.isArray(prData.reports) || prData.reports.length === 0) {
        throw new Error('No coverage reports found in the API');
      }

      const latestReport = prData.reports[0];
      const latestReportCreatedAtMs = parseTimestamp(
        latestReport.created_at,
        `${prLabel} latest report created_at`,
      );

      if (latestReport.status === 'pending') {
        throw new Error('Latest coverage report is still pending');
      }

      if (latestReport.status === 'failure') {
        throw new Error('Latest coverage report failed');
      }

      if (latestReport.status !== 'success') {
        throw new Error(`Latest coverage report has unexpected status "${latestReport.status}"`);
      }

      if (latestReportCreatedAtMs + REPORT_CLOCK_SKEW_MS < codeUpdatedAtMs) {
        throw new Error(
          `Latest successful coverage report at ${latestReport.created_at} predates code update at ${targetPR.code_updated_at}`,
        );
      }

      log.info(
        `API checks passed for ${prLabel}. Latest report created ${latestReport.created_at}, commit ${latestReport.commit}`,
      );

      // Step 3: Open the PR page in the browser to verify the frontend can render the report.
      if (!page) {
        page = await synthetics.getPage();
        await page.setViewport({ width: 1280, height: 800 });
      }

      const { prTitle, accordionCount } = await verifyFrontend(page, targetPR);

      log.info(
        `Health check passed using ${prLabel}. Page title: "${prTitle}". ${accordionCount} coverage file(s) visible on the page.`,
      );
      return;
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      candidateFailures.push(`${prLabel}: ${message}`);
      log.info(`Skipping ${prLabel}. ${message}`);
    }
  }

  throw new Error(
    `No usable coverage report found among ${candidatePRs.length} recent candidate PR(s): ${candidateFailures.join(' | ')}`,
  );
};

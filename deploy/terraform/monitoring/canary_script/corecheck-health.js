const synthetics = require('Synthetics');
const log = require('SyntheticsLogger');

const API_BASE = 'https://api.corecheck.dev';
const FRONTEND_BASE = 'https://corecheck.dev/bitcoin/bitcoin/pulls';

exports.handler = async function () {
  // Step 1: Fetch the PR list and find one updated more than 2 hours ago.
  // A PR updated 2+ hours ago has had enough time for the coverage pipeline to finish.
  log.info('Fetching PR list from API...');
  const listResponse = await fetch(`${API_BASE}/pulls?page=1&title=`);
  if (!listResponse.ok) {
    throw new Error(`Pulls API returned HTTP ${listResponse.status} - API may be down`);
  }

  const pulls = await listResponse.json();
  if (!Array.isArray(pulls) || pulls.length === 0) {
    throw new Error('No PRs returned from API - API or database may be down');
  }

  const twoHoursAgo = Date.now() - 2 * 60 * 60 * 1000;
  const todayStart = new Date();
  todayStart.setHours(0, 0, 0, 0);

  // Find a PR updated between start of today and 2 hours ago so we know it
  // was processed today but has had time to complete.
  const targetPR = pulls.find((pr) => {
    if (!pr.updated_at) return false;
    const updatedAt = new Date(pr.updated_at).getTime();
    return updatedAt >= todayStart.getTime() && updatedAt <= twoHoursAgo;
  });

  if (!targetPR) {
    throw new Error(
      'No PR found that was updated today and at least 2 hours ago - ' +
        'the coverage pipeline may be stalled or no PRs have been processed today',
    );
  }

  log.info(`Using PR #${targetPR.number}: "${targetPR.title}" (updated ${targetPR.updated_at})`);

  // Step 2: Verify the PR has a report via the API and that it is from today.
  const prResponse = await fetch(`${API_BASE}/pulls/${targetPR.number}`);
  if (!prResponse.ok) {
    throw new Error(
      `PR detail API returned HTTP ${prResponse.status} for PR #${targetPR.number}`,
    );
  }

  const prData = await prResponse.json();
  if (!prData.reports || prData.reports.length === 0) {
    throw new Error(`PR #${targetPR.number} has no coverage reports in the API`);
  }

  const latestReport = prData.reports[0];
  const today = new Date().toISOString().split('T')[0];
  const reportDate = new Date(latestReport.created_at).toISOString().split('T')[0];

  if (reportDate !== today) {
    throw new Error(
      `Latest coverage report for PR #${targetPR.number} is from ${reportDate}, ` +
        `expected today (${today}) - pipeline may have stalled`,
    );
  }

  if (latestReport.status === 'pending') {
    throw new Error(
      `Latest coverage report for PR #${targetPR.number} is still pending after 2+ hours`,
    );
  }

  log.info(
    `API checks passed. Report date: ${reportDate}, commit: ${latestReport.commit}, status: ${latestReport.status}`,
  );

  // Step 3: Open the PR page in the browser to verify the frontend is working.
  const page = await synthetics.getPage();
  await page.setViewport({ width: 1280, height: 800 });

  const prUrl = `${FRONTEND_BASE}/${targetPR.number}`;
  log.info(`Navigating to ${prUrl}`);

  const navResponse = await page.goto(prUrl, { waitUntil: 'networkidle2', timeout: 30000 });
  if (!navResponse || navResponse.status() >= 400) {
    throw new Error(
      `Frontend returned HTTP ${navResponse ? navResponse.status() : 'null'} for ${prUrl}`,
    );
  }

  // Step 4: Wait for the PR title heading to confirm the page loaded correctly.
  await page.waitForSelector('h1', { timeout: 15000 });

  const prTitle = await page.evaluate(() => {
    const h1 = document.querySelector('h1');
    return h1 ? h1.textContent.trim() : null;
  });

  if (!prTitle) {
    throw new Error('PR title heading not found - page may not have loaded correctly');
  }

  log.info(`Page loaded. PR title: "${prTitle}"`);

  // Step 5: Verify the Coverage report form field is present.
  await page.waitForSelector('.form-field', { timeout: 15000 });

  const hasCoverageField = await page.evaluate(() => {
    const labels = document.querySelectorAll('.form-field label .txt');
    return Array.from(labels).some((el) => el.textContent.trim() === 'Coverage report');
  });

  if (!hasCoverageField) {
    throw new Error('Coverage report field not found on the PR page');
  }

  log.info('Coverage report field is present');

  // Step 6: Verify coverage data (file accordions) has rendered on the page.
  await page.waitForSelector('.accordion', { timeout: 20000 });

  const accordionCount = await page.evaluate(() => {
    return document.querySelectorAll('.accordion').length;
  });

  if (accordionCount === 0) {
    throw new Error('No coverage file accordions found - report data may be missing');
  }

  log.info(`Health check passed. ${accordionCount} coverage file(s) visible on the page.`);
};

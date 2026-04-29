import { env } from "$env/dynamic/public";

const COVERAGE_TOTAL_LABELS = [
    "Function Coverage",
    "Line Coverage",
    "Region Coverage",
    "Branch Coverage",
];

function extractCoverageTotals(html) {
    const totalsRowMatch = html.match(
        /<tr class='light-row-bold'><td><pre>Totals<\/pre><\/td>(.*?)<\/tr>/s
    );
    if (!totalsRowMatch) {
        return [];
    }

    const totals = [...totalsRowMatch[1].matchAll(/<td class='[^']*'><pre>\s*([^<]+?)\s*<\/pre><\/td>/g)]
        .slice(0, COVERAGE_TOTAL_LABELS.length)
        .map((match, index) => ({
            label: COVERAGE_TOTAL_LABELS[index],
            value: match[1].trim(),
        }));

    if (totals.length !== COVERAGE_TOTAL_LABELS.length) {
        return [];
    }

    return totals;
}

export async function load({ fetch }) {
    try {
        const response = await fetch(`${env.PUBLIC_ENDPOINT}/master-coverage`);
        if (!response.ok) {
            return { report: null, coverageTotals: [] };
        }

        const report = await response.json();
        let coverageTotals = [];

        if (report?.report_url) {
            try {
                const reportResponse = await fetch(report.report_url);
                if (reportResponse.ok) {
                    coverageTotals = extractCoverageTotals(await reportResponse.text());
                }
            } catch (error) {
                console.error(error);
            }
        }

        return {
            report,
            coverageTotals,
        };
    } catch (error) {
        console.error(error);
        return { report: null, coverageTotals: [] };
    }
}

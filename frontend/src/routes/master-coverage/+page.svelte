<svelte:head>
    <title>Master Coverage</title>
</svelte:head>

<script>
    export let data;

    $: report = data.report;

    function formatDate(value) {
        if (!value) return "—";
        const date = new Date(value);
        if (Number.isNaN(date.getTime())) return value;
        return date.toLocaleString();
    }

    function getGeneratedLabel(reportData) {
        if (!reportData) return "—";
        return formatDate(reportData.generated_at || reportData.created_at);
    }
</script>

<div class="master-coverage-page">
    <div class="page-shell">
        <div class="page-header card">
            <div>
                <div class="eyebrow">Master coverage</div>
                <h1>LLVM coverage report for bitcoin/bitcoin master</h1>
                <p class="subtitle">
                    RAW HTML Coverage report generated with `llvm-cov show --format=html`
                </p>
            </div>

            {#if report}
                <div class="meta-grid">
                    <div class="meta-card">
                        <span class="meta-label">Master commit</span>
                        <a
                            class="meta-value txt-mono"
                            href={"https://github.com/bitcoin/bitcoin/commit/" + report.commit}
                            target="_blank"
                            rel="noreferrer"
                        >
                            {report.commit.slice(0, 7)}
                        </a>
                    </div>
                    <div class="meta-card">
                        <span class="meta-label">Generated</span>
                        <span class="meta-value">{getGeneratedLabel(report)}</span>
                    </div>
                    <div class="meta-card">
                        <span class="meta-label">Status</span>
                        <span class="meta-value status-pill status-{report.status}">
                            {report.status}
                        </span>
                    </div>
                </div>
            {/if}
        </div>

        {#if !report}
            <div class="card alert alert-info">
                <i class="ri-information-line" /> No master coverage report is available yet.
            </div>
        {:else if report.status === "pending"}
            <div class="card alert alert-warning">
                <i class="ri-information-line" /> The latest master coverage report is
                being generated.
            </div>
        {:else if report.status === "failure"}
            <div class="card alert alert-danger">
                <i class="ri-information-line" /> Failed to generate the latest master
                coverage report{report.failure_reason ? `: ${report.failure_reason}` : "."}
            </div>
        {:else if !report.report_url}
            <div class="card alert alert-info">
                <i class="ri-information-line" /> Corecheck is preparing the embeddable
                HTML report for the latest master run.
            </div>
        {:else}
            <div class="report-frame card">
                <div class="report-actions">
                    <a
                        href={report.report_url}
                        target="_blank"
                        rel="noreferrer"
                        class="link-primary"
                    >
                        Open report in a new tab
                    </a>
                </div>
                <iframe
                    title="Master coverage report"
                    src={report.report_url}
                    loading="eager"
                />
            </div>
        {/if}
    </div>
</div>

<style lang="scss">
    .master-coverage-page {
        flex: 1 1 auto;
        min-width: 0;
        padding: clamp(12px, 2vw, 24px);
        height: 100%;
        overflow: auto;
    }

    .page-shell {
        display: flex;
        flex-direction: column;
        gap: clamp(16px, 2vw, 20px);
        min-height: 100%;
        min-width: 0;
    }

    .page-header,
    .report-frame {
        background: #fff;
        border-radius: 16px;
        padding: clamp(18px, 2vw, 24px);
        box-shadow: 0 10px 30px rgba(15, 23, 42, 0.08);
    }

    .eyebrow {
        font-size: 0.8rem;
        font-weight: 700;
        letter-spacing: 0.08em;
        text-transform: uppercase;
        color: #2f6fec;
        margin-bottom: 8px;
    }

    h1 {
        margin: 0;
        font-size: 2rem;
        line-height: 1.2;
    }

    .subtitle {
        margin: 12px 0 0;
        color: #5f6b7a;
        max-width: 60rem;
    }

    .meta-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(min(180px, 100%), 1fr));
        gap: 12px;
        margin-top: 24px;
    }

    .meta-card {
        background: #f8fafc;
        border: 1px solid #e2e8f0;
        border-radius: 12px;
        padding: 14px 16px;
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .meta-label {
        font-size: 0.75rem;
        font-weight: 700;
        letter-spacing: 0.04em;
        text-transform: uppercase;
        color: #64748b;
    }

    .meta-value {
        color: #0f172a;
        font-weight: 600;
        word-break: break-word;
    }

    .status-pill {
        display: inline-flex;
        align-items: center;
        width: fit-content;
        padding: 4px 10px;
        border-radius: 999px;
        text-transform: capitalize;
    }

    .status-success {
        background: #dcfce7;
        color: #166534;
    }

    .status-pending {
        background: #fef3c7;
        color: #92400e;
    }

    .status-failure {
        background: #fee2e2;
        color: #b91c1c;
    }

    .report-frame {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 16px;
        min-height: 70vh;
        min-width: 0;
    }

    .report-actions {
        display: flex;
        justify-content: flex-end;
    }

    iframe {
        display: block;
        flex: 1;
        width: 100%;
        min-height: clamp(28rem, 70vh, 64rem);
        border: 1px solid #e2e8f0;
        border-radius: 12px;
        background: #fff;
    }

    @media (max-width: 768px) {
        h1 {
            font-size: clamp(1.35rem, 4vw, 1.6rem);
        }

        .report-actions {
            justify-content: flex-start;
        }

        iframe {
            min-height: 60vh;
        }
    }
</style>

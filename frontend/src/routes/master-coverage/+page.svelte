<svelte:head>
    <title>Master Coverage</title>
</svelte:head>

<script>
    export let data;

    $: report = data.report;
    $: coverageTotals = data.coverageTotals || [];

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
            <div class="title-copy">
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
                </div>
            {/if}
        </div>

        {#if coverageTotals.length}
            <section class="coverage-totals-section" aria-labelledby="coverage-totals-heading">
                <h2 id="coverage-totals-heading" class="section-heading">Totals</h2>
                <div class="coverage-totals">
                    {#each coverageTotals as total}
                        <div class="coverage-total-card">
                            <span class="coverage-total-label">{total.label}</span>
                            <span class="coverage-total-value txt-mono">{total.value}</span>
                        </div>
                    {/each}
                </div>
            </section>
        {/if}

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

    .page-header {
        display: grid;
        grid-template-columns: minmax(0, 1fr) auto;
        align-items: start;
        gap: 16px 24px;
        padding: clamp(16px, 1.75vw, 20px);
    }

    .title-copy {
        min-width: 0;
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
        font-size: clamp(1.65rem, 2.2vw, 1.9rem);
        line-height: 1.2;
    }

    .subtitle {
        margin: 8px 0 0;
        color: #5f6b7a;
        max-width: 60rem;
    }

    .meta-grid {
        display: flex;
        flex-wrap: wrap;
        gap: 12px;
        justify-content: flex-end;
        align-self: start;
    }

    .meta-card {
        background: #f8fafc;
        border: 1px solid #e2e8f0;
        border-radius: 12px;
        padding: 12px 14px;
        display: flex;
        flex-direction: column;
        gap: 6px;
        min-width: min(220px, 100%);
    }

    .coverage-totals {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(min(220px, 100%), 1fr));
        gap: 12px;
    }

    .coverage-totals-section {
        display: flex;
        flex-direction: column;
        gap: 10px;
        margin-top: -4px;
    }

    .section-heading {
        margin: 0;
        font-size: 0.95rem;
        font-weight: 700;
        letter-spacing: 0.04em;
        text-transform: uppercase;
        color: #64748b;
    }

    .coverage-total-card {
        background: #fff;
        border: 1px solid #e2e8f0;
        border-radius: 16px;
        padding: 16px 18px;
        box-shadow: 0 10px 30px rgba(15, 23, 42, 0.08);
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .coverage-total-label {
        font-size: 0.75rem;
        font-weight: 700;
        letter-spacing: 0.04em;
        text-transform: uppercase;
        color: #64748b;
    }

    .coverage-total-value {
        color: #0f172a;
        font-size: 1.05rem;
        font-weight: 600;
        word-break: break-word;
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
        .page-header {
            grid-template-columns: 1fr;
            gap: 14px;
        }

        h1 {
            font-size: clamp(1.35rem, 4vw, 1.55rem);
        }

        .meta-grid {
            justify-content: flex-start;
        }

        .meta-card {
            min-width: 100%;
        }

        .coverage-total-card {
            padding: 14px 16px;
        }

        .report-actions {
            justify-content: flex-start;
        }

        iframe {
            min-height: 60vh;
        }
    }
</style>

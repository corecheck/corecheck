<script lang="ts">
    import { page } from "$app/stores";
    import Field from "../../../../../components/base/Field.svelte";
    import github from "svelte-highlight/styles/github";
    import tooltip from "../../../../../actions/tooltip";
    import Select from "@/components/base/Select.svelte";
    import CoverageReportSelectOption from "./CoverageReportSelectOption.svelte";
    import Coverage from "./Coverage.svelte";
    import { env } from "$env/dynamic/public";
    import { _fetchReport, _fetchSonarCloudIssues } from "@/lib/shared/report";
    import Sonarcloud from "./Sonarcloud.svelte";
    import Benchmarks from "./Benchmarks.svelte";
    const pageTitle = "Pull requests";
    export let data;
    let { pr, sonarcloud, report } = data;
    console.log(report);

    let selectedReport = pr.reports?.[0];
    let prev = selectedReport;

    let fetching = false;
    let showDebugInfo = false;

    function getReportFailureReason(report: { failure_reason?: string } | null | undefined) {
        return (
            report?.failure_reason?.trim() ||
            "The coverage workflow failed for an unknown reason."
        );
    }

    $: {
        (async () => {
            if (selectedReport && selectedReport.id !== prev?.id) {
                prev = selectedReport;
                showDebugInfo = false;
                fetching = true;
                report = await _fetchReport(
                    fetch,
                    env.PUBLIC_ENDPOINT,
                    pr.number,
                    selectedReport.id,
                );

                sonarcloud = await await _fetchSonarCloudIssues(fetch, report.pr_number, report.commit);
                fetching = false;
            }
        })();
    }
</script>

<svelte:head>
    {@html github}
</svelte:head>

<div class="page-wrapper">
    <main class="page-content">
        <div class="page-header-wrapper m-b-0">
            <header class="page-header">
                <nav class="breadcrumbs">
                    <div class="breadcrumb-item">{pageTitle}</div>
                    <div class="breadcrumb-item">#{$page.params.number}</div>
                </nav>
                <div class="flex-fill" />
            </header>
            <h1 class="flex">
                <a
                    href={"https://github.com/bitcoin/bitcoin/pull/" +
                        pr.number}
                    target="_blank"
                    use:tooltip={{
                        text: pr.head.substr(0, 7),
                        position: "right",
                    }}
                >
                    {pr.title}
                </a>
            </h1>

            <div class="clearfix m-b-base" />

            <Field class="form-field" name="coverage" let:uniqueId>
                <label for={uniqueId}>
                    <i class="ri-file-code-line" />
                    <span class="txt">Coverage report</span>
                </label>
                <Select
                    id={uniqueId}
                    toggle={true}
                    items={pr.reports || []}
                    bind:selected={selectedReport}
                    on:change={(e) => console.log(e)}
                    labelComponent={CoverageReportSelectOption}
                    optionComponent={CoverageReportSelectOption}
                />
            </Field>
            {#if report?.base_commit}
                <div class="report-meta">
                    <span class="report-meta-label">Base master commit</span>
                    <a
                        class="label label-sm link-primary txt-mono"
                        href={"https://github.com/bitcoin/bitcoin/commit/" +
                            report.base_commit}
                        target="_blank"
                        rel="noreferrer"
                    >
                        {report.base_commit.substring(0, 7)}
                    </a>
                    <span class="report-meta-description">
                        used for coverage and benchmarks
                    </span>
                </div>
            {/if}
        </div>
        <div class="clearfix m-b-base" />

        {#if fetching}
            <div class="alert alert-info" style="text-align: center">
                <i class="ri-information-line" /> Loading coverage report...
            </div>
        {:else}
            {#if !report}
                <div class="alert alert-info" style="text-align: center">
                    <i class="ri-information-line" /> No coverage data available.
                </div>
            {/if}
            {#if report && report.status === "pending"}
                <div class="alert alert-warning" style="text-align: center">
                    <i class="ri-information-line" /> Coverage report is currently
                    being generated, please come back later.
                </div>
            {/if}
            {#if report && report.status === "failure"}
                <div class="alert alert-danger" style="text-align: center">
                    <i class="ri-information-line" /> Failed to generate report:
                    {getReportFailureReason(report)}
                </div>
            {/if}
            {#if report && report.status === "success"}
                {#if report?.coverage}
                    <div
                        class="flex cov-container flex-justify-between flex-align-start bg-grey"
                    >
                        <div class="cov-col">
                            <Coverage
                                name="Uncovered new code"
                                description="Lines of code added in this pull request that are not covered by tests."
                                coverage={report.coverage.uncovered_new_code}
                                icon="ri-alert-line"
                                color="txt-danger"
                            />
                        </div>
                        <div class="cov-col">
                            <Coverage
                                name="Covered new code"
                                description="Lines of code added in this pull request that are covered by tests."
                                coverage={report.coverage
                                    .gained_coverage_new_code}
                                icon="ri-check-line"
                                color="txt-success"
                            />
                        </div>
                    </div>
                    <div
                        class="flex cov-container flex-justify-between flex-align-start"
                    >
                        <div class="cov-col">
                            <Coverage
                                name="Lost baseline coverage"
                                description="Lines of code that were covered by tests in master but are not covered anymore in this pull request."
                                coverage={report.coverage
                                    .lost_baseline_coverage}
                                icon="ri-alert-line"
                                color="txt-danger"
                            />
                        </div>
                        <div class="cov-col">
                            <Coverage
                                name="Gained baseline coverage"
                                description="Lines of code that were not covered by tests in master but are covered in this pull request."
                                coverage={report.coverage
                                    .gained_baseline_coverage}
                                icon="ri-check-line"
                                color="txt-success"
                            />
                        </div>
                    </div>
                    <div
                        class="flex cov-container flex-justify-between flex-align-start bg-grey"
                    >
                        <div class="cov-col">
                            <Coverage
                                name="Uncovered included code"
                                description="Lines of code that were not executed in master but are executed in this pull request and are not covered by tests."
                                coverage={report.coverage
                                    .uncovered_included_code}
                                icon="ri-alert-line"
                                color="txt-danger"
                            />
                        </div>
                        <div class="cov-col">
                            <Coverage
                                name="Covered included code"
                                description="Lines of code that were not executed in master but are executed in this pull request and are covered by tests."
                                coverage={report.coverage
                                    .gained_coverage_included_code}
                                icon="ri-check-line"
                                color="txt-success"
                            />
                        </div>
                    </div>
                {/if}

                {#if sonarcloud}
                    <div class="cov-col">
                        <Sonarcloud {report} issues={sonarcloud.issues} />
                    </div>
                {/if}
                <div class="clearfix m-b-base" />
                <div class="cov-col full-width">
                    <Benchmarks {report} />
                </div>
            {/if}
        {/if}
        {#if report}
            <div class="debug-info-footer">
                <button
                    class="debug-info-toggle link-primary txt-xs"
                    type="button"
                    aria-controls="report-debug-trace"
                    aria-expanded={showDebugInfo}
                    on:click={() => (showDebugInfo = !showDebugInfo)}
                >
                    {showDebugInfo ? "Hide debug info" : "Debug info"}
                </button>
                <div
                    id="report-debug-trace"
                    class="debug-info-panel"
                    hidden={!showDebugInfo}
                    aria-hidden={!showDebugInfo}
                    data-report-id={report.id}
                    data-report-commit={report.commit}
                    data-step-function-execution-arn={report.step_function_execution_arn ||
                        ""}
                    data-coverage-batch-job-id={report.coverage_batch_job_id || ""}
                >
                    <div class="debug-info-row">
                        <span class="debug-info-label">Report ID</span>
                        <code class="debug-info-value txt-mono">{report.id}</code>
                    </div>
                    <div class="debug-info-row">
                        <span class="debug-info-label">Commit</span>
                        <code class="debug-info-value txt-mono">{report.commit}</code>
                    </div>
                    <div class="debug-info-row">
                        <span class="debug-info-label"
                            >Step Functions execution ARN</span
                        >
                        <code class="debug-info-value txt-mono"
                            >{report.step_function_execution_arn ||
                                "Not available"}</code
                        >
                    </div>
                    <div class="debug-info-row">
                        <span class="debug-info-label">Coverage batch job ID</span>
                        <code class="debug-info-value txt-mono"
                            >{report.coverage_batch_job_id || "Not available"}</code
                        >
                    </div>
                </div>
            </div>
        {/if}
        <div class="clearfix m-b-base" />
    </main>
</div>

<style lang="scss">
    .cov-container {
        @media (max-width: 1100px) {
            flex-direction: column;

            .cov-col {
                width: 100% !important;
            }
        }
        .cov-col {
            width: 48%;
        }
    }

    .report-meta {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        flex-wrap: wrap;
    }

    .report-meta-label {
        font-weight: 600;
    }

    .report-meta-description {
        color: #65717d;
    }

    .debug-info-footer {
        margin-top: 1.5rem;
        padding-top: 0.75rem;
        border-top: 1px solid var(--baseAlt2Color);
    }

    .debug-info-toggle {
        padding: 0;
        border: 0;
        background: none;
        cursor: pointer;
        text-decoration: underline;
    }

    .debug-info-panel {
        margin-top: 0.75rem;
        padding: 0.75rem 1rem;
        background: var(--baseAlt1Color);
        border: 1px solid var(--baseAlt2Color);
    }

    .debug-info-row {
        display: grid;
        grid-template-columns: minmax(0, 220px) minmax(0, 1fr);
        gap: 0.5rem 1rem;
        align-items: start;
    }

    .debug-info-row + .debug-info-row {
        margin-top: 0.5rem;
        padding-top: 0.5rem;
        border-top: 1px solid var(--baseAlt2Color);
    }

    .debug-info-label {
        color: var(--txtHintColor);
        font-size: var(--xsFontSize);
    }

    .debug-info-value {
        display: block;
        white-space: pre-wrap;
        word-break: break-word;
    }

    @media (max-width: 700px) {
        .debug-info-row {
            grid-template-columns: 1fr;
        }
    }
</style>

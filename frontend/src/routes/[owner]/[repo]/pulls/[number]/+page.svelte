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

    let selectedReport = pr.reports[0];
    let prev = selectedReport;

    let fetching = false;

    $: {
        (async () => {
            if (selectedReport.id !== prev.id) {
                prev = selectedReport;
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
                    <i class="ri-information-line" /> An error occured while generating
                    the coverage report.
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
</style>

<script lang="ts">
    import Accordion from "@/components/base/Accordion.svelte";
    import { page } from "$app/stores";

    export let report: any;
    export let issues: any = [];

</script>
<h1>Sonarcloud</h1>
<div class="clearfix m-b-base" />

<div class="accordions sonarcloud">
    {#each issues.sort((a, b) => {
        if (a.severity === "BLOCKER") return -1;
        if (b.severity === "BLOCKER") return 1;
        if (a.severity === "CRITICAL") return -1;
        if (b.severity === "CRITICAL") return 1;
        if (a.severity === "MAJOR") return -1;
        if (b.severity === "MAJOR") return 1;
        if (a.severity === "MINOR") return -1;
        if (b.severity === "MINOR") return 1;
        if (a.severity === "INFO") return -1;
        if (b.severity === "INFO") return 1;
        return 0;
    }) as issue}
        <Accordion>
            <svelte:fragment slot="header">
                <div class="inline-flex">
                    <i class="ri-file-code-line" />
                    <span class="txt">
                        {issue.message}
                    </span>
                </div>

                <div class="flex-fill" />
                <span
                    class="label"
                    class:label-success={issue.severity === "BLOCKER"}
                    class:label-warning={issue.severity === "CRITICAL"}
                    class:label-danger={issue.severity === "MAJOR"}
                    class:label-info={issue.severity === "MINOR"}
                    class:label-secondary={issue.severity === "INFO"}
                    >{issue.severity}</span
                >
            </svelte:fragment>
            <div class="form-field m-b-0">
                <pre><div class="code"><span class="filename"
                            >{issue.component.split(":")[1]}</span
                        >
{#each issue.sources as line}<div class="line"><span
                                    class="line-number txt-mono"
                                    >{line.line} </span
                                ><span
                                    class="txt-mono"
                                    class:highlight={line.line >=
                                        issue.textRange.startLine &&
                                        line.line <= issue.textRange.endLine}
                                    >{@html line.code}</span
                                ></div>{/each}</div></pre>

                <!-- https://sonarcloud.io/project/issues?&sinceLeakPeriod=true&branch=26415&id=aureleoules_bitcoin&open=AYsFGxvQ890w8U3JDlEV -->
                <a
                    class="btn btn-primary btn-sm"
                    href="https://sonarcloud.io/project/issues?id=aureleoules_bitcoin&branch={report.pr_number}-{report.commit}&resolved=false&open={issue.key}"
                    target="_blank">Open in SonarCloud</a
                >
            </div>
        </Accordion>
    {/each}
</div>

<style>
    .highlight {
        background-color: #e97373;
    }
    .full-width {
        width: 100% !important;
    }

    .code {
        background-color: #f1f1f1;
        padding: 1rem;
        border-radius: 0.25rem;
        margin-bottom: 1rem;
        overflow: auto;
    }

    .line {
        display: flex;
        align-items: center;
    }
    .line-number {
        text-align: right;
        margin-right: 0.5rem;
        border-right: 1px solid #d0d5db;
    }

    :global(pre code.hljs) {
        white-space: pre-wrap;
    }
</style>

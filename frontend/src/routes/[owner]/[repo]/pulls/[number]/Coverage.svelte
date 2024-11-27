<script lang="ts">
    import tooltip from "@/actions/tooltip";
    import Accordion from "@/components/base/Accordion.svelte";

    export let coverage: any;
    export let name: string;
    export let description: string;
    export let icon: string;
    export let color: string;

    // randomKey is used to differenciate accordions between Coverage components
    const randomKey = Math.random()
        .toString(36)
        .replace(/[^a-z]+/g, "")
        .substring(0, 5);
    let hasExpanded = false;

    function collapse() {
        const accordionsDiv = document.querySelector(
            ".accordions." + randomKey,
        );
        const accordions = accordionsDiv?.querySelectorAll(".accordion");
        for (const accordion of accordions) {
            if (!accordion.classList.contains("active")) {
                continue;
            }

            accordion.querySelector(".accordion-header").click();
        }

        hasExpanded = false;
    }

    function expand() {
        const accordionsDiv = document.querySelector(
            ".accordions." + randomKey,
        );
        const accordions = accordionsDiv?.querySelectorAll(".accordion");
        for (const accordion of accordions) {
            if (accordion.classList.contains("active")) {
                continue;
            }

            accordion.querySelector(".accordion-header").click();
        }

        hasExpanded = true;
    }

    function getLineTooltip(line) {
        if (!line.tested) return "Not executed";

        if (line.covered && line.highlight) return "This line is covered";
        if (line.covered && line.context) return "This line is covered";
        if (!line.covered && line.highlight) return "This line is not covered";
        if (!line.covered && line.context) return "This line is not covered";

        return "Not executed";
    }
</script>

<div class="flex">
    <div>
        <div class="flex">
            <h1>{name} <i class={`${color} ${icon}`} /></h1>
            {#if coverage}
                {#if hasExpanded}
                    <button
                        type="button"
                        class="btn btn-sm btn-primary"
                        on:click={collapse}
                    >
                        <i class="ri-arrow-up-s-line" />
                        Collapse all
                    </button>
                {:else}
                    <button
                        type="button"
                        class="btn btn-sm btn-primary"
                        on:click={expand}
                    >
                        <i class="ri-arrow-down-s-line" />
                        Expand all
                    </button>
                {/if}
            {/if}
        </div>
        <p>{description}</p>
    </div>
</div>
<div class="clearfix m-b-base" />
{#if !coverage}
    <div class="alert alert-info" style="text-align: center">
        <i class="ri-information-line" /> No coverage data
    </div>
{:else}
    <div class={`accordions ${randomKey}`}>
        {#each Object.keys(coverage) as filename}
            <Accordion>
                <svelte:fragment slot="header">
                    <div class="inline-flex">
                        <i class="ri-file-code-line" />
                        <span class="txt">{filename}</span>
                    </div>

                    <div class="flex-fill" />
                </svelte:fragment>
                {#each coverage[filename] as hunk}
                    <pre><div class="code">{#each hunk.lines as line}<div
                                    class="line"><a
                                        target="_blank"
                                        class="line-number link-primary txt-mono"
                                        >{line.line_number} </a><span
                                        use:tooltip={{
                                            text: getLineTooltip(line),
                                            position: "top",
                                        }}
                                        class:line-changed-covered={line.tested &&
                                            line.highlight &&
                                            line.covered}
                                        class:line-unchanged-covered={line.tested &&
                                            line.covered &&
                                            line.context}
                                        class:line-changed-uncovered={line.tested &&
                                            line.highlight &&
                                            !line.covered}
                                        class:line-unchanged-uncovered={line.tested &&
                                            !line.covered &&
                                            line.context}
                                        class="txt-mono">{line.content}</span
                                    ></div>{/each}</div></pre>
                {/each}
            </Accordion>
        {/each}
    </div>
{/if}
<div class="clearfix m-b-base" />

<style lang="scss">
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

    .filename {
        margin-bottom: 0.5rem;
        background-color: #2f2f30;
        color: #fff;
        padding: 0.1rem 0.5rem;
        font-size: 0.8rem;
        border-radius: 0.5rem;
        display: inline-block;
    }

    .line {
        display: flex;
        align-items: center;

        .line-changed-covered {
            background-color: #66ee86;
        }

        .line-changed-uncovered {
            background-color: #e97373;
        }

        .line-unchanged-covered {
            background-color: #bbf1c8;
        }

        .line-unchanged-uncovered {
            background-color: #f1c7c7;
        }
    }
    .line-number {
        text-align: right;
        margin-right: 0.5rem;
        border-right: 1px solid #d0d5db;
    }

    .context-button {
        position: absolute;
        top: 0px;
        right: 15px;
    }

    :global(pre code.hljs) {
        white-space: pre-wrap;
    }
</style>

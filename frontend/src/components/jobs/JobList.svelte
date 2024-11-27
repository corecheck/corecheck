<script>
    import { createEventDispatcher } from "svelte";
    import CommonHelper from "@/utils/CommonHelper";
    import SortHeader from "@/components/base/SortHeader.svelte";
    import FormattedDate from "@/components/base/FormattedDate.svelte";
    import HorizontalScroller from "@/components/base/HorizontalScroller.svelte";
    import dayjs from "dayjs";

    export let sort = "-rowid";
    export let jobs;
</script>

<HorizontalScroller class="table-wrapper">
    <table class="table">
        <thead>
            <tr>
                <SortHeader disable class="col-type-number col-field-id" name="id" bind:sort>
                    <div class="col-header-content">
                        <i class="ri-key-line" />
                        <span class="txt">ID</span>
                    </div>
                </SortHeader>
                <SortHeader class="col-type-text col-field-type" name="type" bind:sort>
                    <div class="col-header-content">
                        <i class={CommonHelper.getFieldTypeIcon("text")} />
                        <span class="txt">Type</span>
                    </div>
                </SortHeader>
                <SortHeader class="col-type-number col-field-pr" name="pr" bind:sort>
                    <div class="col-header-content">
                        <i class={CommonHelper.getFieldTypeIcon("number")} />
                        <span class="txt">Pull request</span>
                    </div>
                </SortHeader>
                <SortHeader disable class="col-field-status" name="status" bind:sort>
                    <div class="col-header-content">
                        <i class={CommonHelper.getFieldTypeIcon("status")} />
                        <span class="txt">Status</span>
                    </div>
                </SortHeader>

                <SortHeader disable class="col-type-date col-field-created min-width" name="created" bind:sort>
                    <div class="col-header-content">
                        <i class={CommonHelper.getFieldTypeIcon("date")} />
                        <span class="txt">Created at</span>
                    </div>
                </SortHeader>
                <SortHeader disable class="col-type-date col-field-started min-width" name="started" bind:sort>
                    <div class="col-header-content">
                        <i class={CommonHelper.getFieldTypeIcon("date")} />
                        <span class="txt">Started at</span>
                    </div>
                </SortHeader>
                <th class="col-type-action min-width" />
            </tr>
        </thead>
        <tbody>
            {#each jobs as item (item.id)}
                <tr
                    tabindex="0"
                    class="row-handle"
                >
                    <td class="col-type-text col-field-id min-width">
                        <span class="label">
                            {item.id}
                        </span>
                    </td>
                    <td class="col-type-number col-field-type min-width">
                        <span class="label" class:label-success={item.type === "code-coverage"}>
                            {item.type}
                        </span>
                    </td>
                    <td class="col-type-number col-field-id min-width">
                        <span class="label">
                            #{item.pull_request_number}
                        </span>
                    </td>
                    <td class="col-type-text col-field-status">
                        <span class="txt txt-ellipsis" title={item.stat}>
                            {item.aws_status}
                        </span>
                    </td>

                    <td class="col-type-date col-field-created">
                        <span class="txt txt-ellipsis" title={item.created_at}>
                            {dayjs(item.created_at).format("YYYY-MM-DD HH:mm:ss")}
                        </span>
                    </td>
                    <td class="col-type-date col-field-started">
                        <span class="txt txt-ellipsis" title={item.started_at}>
                            {dayjs(item.started_at).format("YYYY-MM-DD HH:mm:ss")}
                        </span>
                    </td>
                </tr>
            {:else}
                <tr>
                    <td colspan="99" class="txt-center txt-hint p-xs">
                        <h6>No jobs found.</h6>
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>
</HorizontalScroller>

{#if jobs.length}
    <small class="block txt-hint txt-right m-t-sm">Showing {jobs.length}</small>
{/if}

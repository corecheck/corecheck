<script>
    import CommonHelper from "@/utils/CommonHelper";
    import SortHeader from "@/components/base/SortHeader.svelte";
    import FormattedDate from "@/components/base/FormattedDate.svelte";
    import HorizontalScroller from "@/components/base/HorizontalScroller.svelte";

    export let filter = "";
    export let presets = "";
    export let sort = "-rowid";

    export let items = [];
    let currentPage = 1;
    let totalItems = 0;
    let isLoading = false;
</script>

<HorizontalScroller class="table-wrapper">
    <table class="table" class:table-loading={isLoading}>
        <thead>
            <tr>
                <SortHeader disable class="col-type-number col-field-status" name="status" bind:sort>
                    <div class="col-header-content">
                        <i class={CommonHelper.getFieldTypeIcon("status")} />
                        <span class="txt">Status</span>
                    </div>
                </SortHeader>
                <SortHeader disable class="col-type-number col-field-number" name="number" bind:sort>
                    <div class="col-header-content">
                        <i class={CommonHelper.getFieldTypeIcon("number")} />
                        <span class="txt">Number</span>
                    </div>
                </SortHeader>
                <SortHeader class="col-type-text col-field-url" name="title" bind:sort>
                    <div class="col-header-content">
                        <i class={CommonHelper.getFieldTypeIcon("text")} />
                        <span class="txt">Title</span>
                    </div>
                </SortHeader>
                <SortHeader disable class="col-field-author" name="author" bind:sort>
                    <div class="col-header-content">
                        <i class="ri-global-line" />
                        <span class="txt">Author</span>
                    </div>
                </SortHeader>

                <SortHeader disable class="col-type-date col-field-created" name="created" bind:sort>
                    <div class="col-header-content">
                        <i class={CommonHelper.getFieldTypeIcon("date")} />
                        <span class="txt">Created</span>
                    </div>
                </SortHeader>
                <SortHeader disable class="col-type-date col-field-updated" name="created" bind:sort>
                    <div class="col-header-content">
                        <i class={CommonHelper.getFieldTypeIcon("date")} />
                        <span class="txt">Updated</span>
                    </div>
                </SortHeader>

                <th class="col-type-action min-width" />
            </tr>
        </thead>
        <tbody>
            {#each items as item}
                <tr
                    tabindex="0"
                    class="row-handle"
                    on:click={() => window.open("/bitcoin/bitcoin/pulls/" + item.number, "_blank")}
                >
                    <td class="col-type-text col-field-method min-width">
                        <span class="label" class:label-danger={item.state === "closed"} class:label-success={item.state === "open"}>
                            {item.state}
                        </span>
                    </td>
                    <td class="col-type-number col-field-number min-width">
                        <span class="label" class:label-danger={item.status >= 400}>
                            {item.number}
                        </span>
                    </td>
                    <td class="col-type-text col-field-url">
                        <span class="txt txt-ellipsis" title={item.title}>
                            {item.title}
                        </span>
                    </td>

                    <td class="col-type-text col-field-method">
                        <span class="label">
                            {item.user}
                        </span>
                    </td>
                    
                    <td class="col-type-date col-field-created">
                        <FormattedDate date={item.created_at} />
                    </td>
                    <td class="col-type-date col-field-created">
                        <FormattedDate date={item.updated_at} />
                    </td>

                    <td class="col-type-action min-width">
                        <i class="ri-arrow-right-line" />
                    </td>
                </tr>
            {:else}
                {#if isLoading}
                    <tr>
                        <td colspan="99" class="p-xs">
                            <span class="skeleton-loader m-0" />
                        </td>
                    </tr>
                {:else}
                    <tr>
                        <td colspan="99" class="txt-center txt-hint p-xs">
                            <h6>No logs found.</h6>
                            {#if filter?.length}
                                <button
                                    type="button"
                                    class="btn btn-hint btn-expanded m-t-sm"
                                    on:click={() => (filter = "")}
                                >
                                    <span class="txt">Clear filters</span>
                                </button>
                            {/if}
                        </td>
                    </tr>
                {/if}
            {/each}
        </tbody>
    </table>
</HorizontalScroller>
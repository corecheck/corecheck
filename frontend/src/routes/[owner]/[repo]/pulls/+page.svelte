<script>
    import RefreshButton from "../../../../components/base/RefreshButton.svelte";
    import Searchbar from "../../../../components/base/Searchbar.svelte";
    import PRList from "@/components/pull-requests/PRList.svelte";

    const pageTitle = "Pull requests";

    export let data;
    let { pulls } = data;

    let refreshKey = 1;
    function refresh() {
        refreshKey++;
    }
</script>

<div class="page-wrapper">
    <main class="page-content">
        <div class="page-header-wrapper m-b-0">
            <header class="page-header">
                <nav class="breadcrumbs">
                    <div class="breadcrumb-item">{pageTitle}</div>
                </nav>
                <RefreshButton on:refresh={() => refresh()} />
                <small style="opacity: 0.6">Updated every hour</small>
                <div class="flex-fill" />
            </header>

            <Searchbar placeholder="Search for PR title or number" />

            <div class="clearfix m-b-base" />
        </div>

        {#key refreshKey}
            <PRList items={pulls} />
        {/key}
    </main>
</div>

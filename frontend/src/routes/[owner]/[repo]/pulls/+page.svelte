<script>
    import RefreshButton from "../../../../components/base/RefreshButton.svelte";
    import Searchbar from "../../../../components/base/Searchbar.svelte";
    import PRList from "@/components/pull-requests/PRList.svelte";
    import { _fetchPulls } from "./+page";

    const pageTitle = "Pull requests";

    export let data;
    let pulls = data.pulls ?? [];
    let searchValue = "";
    let isLoading = false;

    let refreshKey = 1;
    function refresh() {
        refreshKey++;
    }

    async function onSearch(term) {
        searchValue = term ?? "";
        isLoading = true;
        try {
            const result = await _fetchPulls(searchValue, 1);
            pulls = Array.isArray(result) ? result : [];
        } finally {
            isLoading = false;
            refreshKey++;
        }
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

            <Searchbar placeholder="Search for PR title or number" bind:value={searchValue} on:submit={(e) => onSearch(e.detail)} on:clear={() => onSearch("")} />

            <div class="clearfix m-b-base" />
        </div>

        {#key refreshKey}
            <PRList items={pulls} isLoading={isLoading} />
        {/key}
    </main>
</div>

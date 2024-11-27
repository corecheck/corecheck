<script>
    import { createEventDispatcher, onMount } from "svelte";
    import { fly } from "svelte/transition";
    import CommonHelper from "@/utils/CommonHelper";

    const dispatch = createEventDispatcher();
    const uniqueId = "search_" + CommonHelper.randomString(7);

    export let value = "";
    export let placeholder = 'Search term or filter like created > "2022-01-01"...';

    let searchInput;
    let tempValue = "";

    $: if (typeof value === "string") {
        tempValue = value;
    }

    function clear(focusInput = true) {
        tempValue = "";
        if (focusInput) {
            searchInput?.focus();
        }
        dispatch("clear");
    }

    function submit() {
        value = tempValue;
        dispatch("submit", value);
    }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<form class="searchbar" on:click|stopPropagation on:submit|preventDefault={submit}>
    <label for={uniqueId} class="m-l-10 txt-xl">
        <i class="ri-search-line" />
    </label>

    <input
        bind:this={searchInput}
        type="text"
        id={uniqueId}
        placeholder={value || placeholder}
        bind:value={tempValue}
    />

    {#if (value.length || tempValue.length) && tempValue != value}
        <button
            type="submit"
            class="btn btn-expanded btn-sm btn-warning"
            transition:fly|local={{ duration: 150, x: 5 }}
        >
            <span class="txt">Search</span>
        </button>
    {/if}

    {#if value.length || tempValue.length}
        <button
            type="button"
            class="btn btn-transparent btn-sm btn-hint p-l-xs p-r-xs m-l-10"
            transition:fly|local={{ duration: 150, x: 5 }}
            on:click={() => {
                clear(false);
                submit();
            }}
        >
            <span class="txt">Clear</span>
        </button>
    {/if}
</form>

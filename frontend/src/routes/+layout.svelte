<script>
    import Toasts from "../components/base/Toasts.svelte";
    import tooltip from "@/actions/tooltip";
    import { page } from "$app/stores";
    import Toggler from "../components/base/Toggler.svelte";
    import "../scss/main.scss";
    import { setContext } from 'svelte';

    export let data;
    let { user } = data;

    setContext('user', user);
</script>

<div class="app-layout">
    <aside class="app-sidebar">
        <a data-sveltekit-preload-code="eager" href="/" class="logo logo-sm">
            <img
                src="{import.meta.env.BASE_URL}images/logo.png"
                alt="Bitcoin Core Coverage"
                width="40"
                height="40"
            />
        </a>

        <nav data-sveltekit-preload-data="hover" class="main-menu">
            <a
                href="/bitcoin/bitcoin/pulls"
                class="menu-item"
                aria-label="Pull requests"
                class:active={$page.url.pathname.includes("/pulls")}
                use:tooltip={{ text: "Pull requests", position: "right" }}
            >
                <img
                    style="text-align: center"
                    height="25"
                    alt="pr"
                    src="/icons/svg/pull-request.svg"
                />
            </a>
            <a
                data-sveltekit-preload-code="eager"
                href="/tests"
                class="menu-item"
                aria-label="Tests"
                class:active={$page.url.pathname.includes("/tests")}
                use:tooltip={{ text: "Tests", position: "right" }}
            >
                <img
                    style="text-align: center"
                    height="25"
                    alt="jobs"
                    src="/icons/svg/ci.svg"
                />
            </a>
            <a
                data-sveltekit-preload-code="eager"
                href="/benchmarks"
                class="menu-item"
                aria-label="Benchmarks"
                class:active={$page.url.pathname.includes("/benchmarks")}
                use:tooltip={{ text: "Benchmarks", position: "right" }}
            >
                <img
                    style="text-align: center"
                    height="25"
                    alt="jobs"
                    src="/icons/svg/stats.svg"
                />
            </a>
            <a
                data-sveltekit-preload-code="eager"
                href="/jobs"
                class="menu-item"
                aria-label="Jobs"
                class:active={$page.url.pathname.includes("/jobs")}
                use:tooltip={{ text: "Jobs", position: "right" }}
            >
                <img
                    style="text-align: center"
                    height="25"
                    alt="jobs"
                    src="/icons/svg/activity.svg"
                />
            </a>
        </nav>

        {#key user}
        <a
            href="https://github.com/corecheck/corecheck"
            target="_blank"
            rel="noopener noreferrer"
            class="menu-item"
        >
            <i style="font-size: 50px; margin-bottom: 10px" class="ri-github-fill" />
        </a>
        {/key}
    </aside>

    <div class="app-body">
        <slot />
        <Toasts />
    </div>
</div>

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
        <a
            data-sveltekit-preload-code="eager"
            href="/"
            class="menu-item"
            aria-label="GitHub overview"
            class:active={$page.url.pathname === "/"}
            use:tooltip={{ text: "GitHub overview", position: "right" }}
        >
            <span class="menu-item-icon">
                <img
                    src="{import.meta.env.BASE_URL}images/logo.png"
                    alt="GitHub overview"
                    width="40"
                    height="40"
                />
            </span>
            <span class="menu-item-label">GitHub overview</span>
        </a>

        <nav data-sveltekit-preload-data="hover" class="main-menu">
            <a
                href="/bitcoin/bitcoin/pulls"
                class="menu-item"
                aria-label="PR coverage"
                class:active={$page.url.pathname.includes("/pulls")}
                use:tooltip={{ text: "PR coverage", position: "right" }}
            >
                <span class="menu-item-icon">
                    <img height="25" alt="PR coverage" src="/icons/svg/pull-request.svg" />
                </span>
                <span class="menu-item-label">PR coverage</span>
            </a>
            <a
                data-sveltekit-preload-code="eager"
                href="/master-coverage"
                class="menu-item"
                aria-label="Master coverage"
                class:active={$page.url.pathname.includes("/master-coverage")}
                use:tooltip={{ text: "Master coverage", position: "right" }}
            >
                <span class="menu-item-icon">
                    <i class="ri-file-chart-line" />
                </span>
                <span class="menu-item-label">Master coverage</span>
            </a>
            <a
                data-sveltekit-preload-code="eager"
                href="/tests"
                class="menu-item"
                aria-label="Tests"
                class:active={$page.url.pathname.includes("/tests")}
                use:tooltip={{ text: "Tests", position: "right" }}
            >
                <span class="menu-item-icon">
                    <img height="25" alt="Tests" src="/icons/svg/ci.svg" />
                </span>
                <span class="menu-item-label">Tests</span>
            </a>
            <a
                data-sveltekit-preload-code="eager"
                href="/benchmarks"
                class="menu-item"
                aria-label="Benchmarks"
                class:active={$page.url.pathname.includes("/benchmarks")}
                use:tooltip={{ text: "Benchmarks", position: "right" }}
            >
                <span class="menu-item-icon">
                    <img height="25" alt="Benchmarks" src="/icons/svg/stats.svg" />
                </span>
                <span class="menu-item-label">Benchmarks</span>
            </a>
            <a
                data-sveltekit-preload-code="eager"
                href="/mutation"
                class="menu-item"
                aria-label="Mutation"
                class:active={$page.url.pathname.includes("/mutation")}
                use:tooltip={{ text: "Mutation", position: "right" }}
            >
                <span class="menu-item-icon">
                    <img height="25" alt="Mutations" src="/icons/png/mutation.png" />
                </span>
                <span class="menu-item-label">Mutation</span>
            </a>
        </nav>

        <div class="sidebar-footer-links">
            <a
                data-sveltekit-preload-code="eager"
                href="/jobs"
                class="menu-item"
                aria-label="Site health"
                class:active={$page.url.pathname.includes("/jobs")}
                use:tooltip={{ text: "Site health", position: "right" }}
            >
                <span class="menu-item-icon">
                    <img height="25" alt="Site health" src="/icons/svg/activity.svg" />
                </span>
                <span class="menu-item-label">Site health</span>
            </a>

            {#key user}
            <a
                href="https://github.com/corecheck/corecheck"
                target="_blank"
                rel="noopener noreferrer"
                class="menu-item menu-item-icon-only"
                aria-label="GitHub repo"
                use:tooltip={{ text: "GitHub repo", position: "right" }}
            >
                <span class="menu-item-icon">
                    <i class="ri-github-fill" />
                </span>
            </a>
            {/key}
        </div>
    </aside>

    <div class="app-body">
        <slot />
        <Toasts />
    </div>
</div>

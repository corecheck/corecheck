<script>
    import { onMount, tick } from "svelte";
    import { page } from "$app/stores";
    import { get as getStore } from "svelte/store";
    import { goto } from "$app/navigation";
    import { env } from "$env/dynamic/public";

    let selectedFile = "";
    let fileContent = "";
    let mutationData = {};
    let expandedLines = new Set();
    let expanded = new Set(["src", "src/script", "src/wallet"]);
    let highlightedLine = null;

    const GITHUB_RAW_BASE =
        "https://raw.githubusercontent.com/bitcoin/bitcoin/";

    let mutationsMeta = {};
    let mutations = [];
    let countMutations = 0;

    // ===== Vertical-only scroll helper =====
    function scrollLineIntoView(
        lineNumber,
        { alignLeft = false, vertical = "center", smooth = false } = {},
    ) {
        const el = document.getElementById(`L${lineNumber}`);
        if (!el) return;

        const container = document.querySelector(".code-container"); // horizontal scroller

        const rect = el.getBoundingClientRect();
        const currentY = window.scrollY;
        let targetTop;

        if (vertical === "start") {
            const offset = 0; // adjust if you add a fixed header
            targetTop = currentY + rect.top - offset;
        } else if (vertical === "end") {
            targetTop = currentY + rect.bottom - window.innerHeight;
        } else {
            targetTop =
                currentY + rect.top - window.innerHeight / 2 + rect.height / 2;
        }

        window.scrollTo({
            top: targetTop,
            left: window.scrollX, // DO NOT change X here
            behavior: smooth ? "smooth" : "auto",
        });

        if (container && alignLeft) {
            container.scrollLeft = 0; // show start of the line
        }
    }

    // Accepts optional lineToFocus
    async function handleFileSelect(
        file,
        updateUrl = true,
        lineToFocus = null,
    ) {
        selectedFile = file;

        if (updateUrl) {
            const currentHash =
                typeof window !== "undefined" ? window.location.hash : "";
            goto(`/mutation/${file}${currentHash}`, { replaceState: false });
        }

        try {
            // mutations may still be loading on first paint; this is fine
            const selected_mutations = mutations.filter((val) =>
                val.filename.includes(file),
            );
            mutationData =
                selected_mutations[0] && selected_mutations[0].diffs
                    ? selected_mutations[0].diffs
                    : {};

            // Fetch file content
            const commit = selected_mutations[0]?.diffs
                ? (Object.values(selected_mutations[0].diffs)[0]?.[0]?.commit ??
                  "master")
                : "master";
            console.log(commit);
            const githubPath = `${GITHUB_RAW_BASE}/${commit}/${file}`;
            const contentResp = await fetch(githubPath);
            if (!contentResp.ok)
                throw new Error(`Failed to fetch file: ${contentResp.status}`);
            fileContent = await contentResp.text();

            await tick(); // ensure lines are in the DOM

            // Prefer explicit param; otherwise fall back to URL hash (handles first-load race)
            let focus = lineToFocus;
            if (focus == null && typeof window !== "undefined") {
                const m = window.location.hash.match(/^#?L(\d+)$/);
                focus = m ? Number(m[1]) : null;
            }

            if (focus != null) {
                highlightedLine = focus;
                scrollLineIntoView(focus, {
                    alignLeft: true,
                    vertical: "center",
                });
                if (mutationData[focus]) {
                    expandedLines = new Set(expandedLines).add(focus);
                }
            } else {
                highlightedLine = null;
            }
        } catch (error) {
            console.error("Error:", error);
            fileContent = `Error loading content: ${error.message}`;
            mutationData = {};
        }
    }

    function toggleLine(lineNumber) {
        expandedLines = new Set(expandedLines);
        if (expandedLines.has(lineNumber)) expandedLines.delete(lineNumber);
        else expandedLines.add(lineNumber);
    }

    function onFileClick(file) {
        handleFileSelect(file, true);
    }

    const files = {
        src: {
            wallet: { "coinselection.cpp": "content" },
            script: {
                "interpreter.cpp": "content",
                "descriptor.cpp": "content",
            },
            consensus: {
                "tx_verify.cpp": "content",
                "tx_check.cpp": "content",
                "merkle.cpp": "content",
            },
            util: { "asmap.cpp": "content" },
            "pow.cpp": "content",
            "addrman.cpp": "content",
            "txgraph.cpp": "content",
        },
    };

    function fileExists(filePath) {
        const parts = filePath.split("/");
        let current = files;
        for (const part of parts) {
            if (current && typeof current === "object" && part in current)
                current = current[part];
            else return false;
        }
        return true;
    }

    function renderTree(tree, path = "") {
        return Object.entries(tree).map(([name, value]) => {
            const currentPath = path ? `${path}/${name}` : name;
            if (typeof value === "string") {
                return { type: "file", name, path: currentPath };
            } else {
                return {
                    type: "directory",
                    name,
                    path: currentPath,
                    children: renderTree(value, currentPath),
                };
            }
        });
    }
    $: treeData = renderTree(files);

    onMount(async () => {
        try {
            const [metaResp, mutsResp] = await Promise.allSettled([
                fetch(env.PUBLIC_ENDPOINT + "/mutations/meta"),
                fetch(env.PUBLIC_ENDPOINT + "/mutations"),
            ]);

            if (metaResp.status === "fulfilled") {
                mutationsMeta = await metaResp.value.json();
            } else {
                console.error(
                    "Failed to fetch mutation metadata:",
                    metaResp.reason,
                );
            }

            if (mutsResp.status === "fulfilled") {
                mutations = await mutsResp.value.json();
            } else {
                console.error("Failed to fetch mutations:", mutsResp.reason);
            }

            // compute count safely
            if (Array.isArray(mutations)) {
                countMutations = mutations.reduce((total, file) => {
                    if (!file?.diffs) return total;
                    return (
                        total +
                        Object.values(file.diffs).reduce(
                            (sum, diffArray) => sum + diffArray.length,
                            0,
                        )
                    );
                }, 0);
            }
        } catch (e) {
            console.error("Initial data fetch failed:", e);
        }

        {
            const { pathname } = getStore(page).url;
            if (pathname.startsWith("/mutation/")) {
                const filePath = pathname.replace("/mutation/", "");
                const m = (
                    typeof window !== "undefined" ? window.location.hash : ""
                ).match(/^#?L(\d+)$/);
                const targetLine = m ? Number(m[1]) : null;

                if (fileExists(filePath)) {
                    // Don't update URL here; we're responding to current URL
                    await handleFileSelect(filePath, false, targetLine);
                }
            }
        }

        const unsubscribe = page.subscribe(async ($page) => {
            const { pathname } = $page.url;
            // If user navigates to a different file after mount
            if (pathname.startsWith("/mutation/")) {
                const filePath = pathname.replace("/mutation/", "");
                if (fileExists(filePath) && selectedFile !== filePath) {
                    const m = (
                        typeof window !== "undefined"
                            ? window.location.hash
                            : ""
                    ).match(/^#?L(\d+)$/);
                    const targetLine = m ? Number(m[1]) : null;
                    await handleFileSelect(filePath, false, targetLine);
                }
            }
        });

        return unsubscribe;
    });
</script>

<div class="page-wrapper">
    <!-- Sidebar -->
    <aside class="file-tree">
        <h2 class="text-xl font-bold mb-4">Files</h2>
        {#each treeData as item}
            {#if item.type === "file"}
                <div
                    class="pl-6 py-1 cursor-pointer hover:bg-gray-200 {selectedFile ===
                    item.path
                        ? 'bg-blue-100'
                        : ''}"
                    on:click={() => handleFileSelect(item.path)}
                >
                    📄 {item.name}
                </div>
            {:else}
                <div>
                    <div
                        class="py-1 cursor-pointer hover:bg-gray-200 flex items-center"
                        on:click={() => {
                            expanded = new Set(expanded);
                            if (expanded.has(item.path))
                                expanded.delete(item.path);
                            else expanded.add(item.path);
                        }}
                    >
                        <span class="mr-2"
                            >{expanded.has(item.path) ? "📂" : "📁"}</span
                        >
                        {item.name}
                    </div>
                    {#if expanded.has(item.path)}
                        <div class="directory">
                            {#each item.children as child}
                                {#if child.type === "file"}
                                    <div
                                        class="pl-6 py-1 cursor-pointer hover:bg-gray-200 {selectedFile ===
                                        child.path
                                            ? 'bg-blue-100'
                                            : ''}"
                                        on:click={() =>
                                            handleFileSelect(child.path)}
                                    >
                                        📄 {child.name}
                                    </div>
                                {:else}
                                    <div>
                                        <div
                                            class="py-1 cursor-pointer hover:bg-gray-200 flex items-center"
                                            on:click={() => {
                                                expanded = new Set(expanded);
                                                if (expanded.has(child.path))
                                                    expanded.delete(child.path);
                                                else expanded.add(child.path);
                                            }}
                                        >
                                            <span class="mr-2"
                                                >{expanded.has(child.path)
                                                    ? "📂"
                                                    : "📁"}</span
                                            >
                                            {child.name}
                                        </div>
                                        {#if expanded.has(child.path)}
                                            <div class="directory">
                                                {#each child.children as grandChild}
                                                    <div
                                                        class="pl-6 py-1 cursor-pointer hover:bg-gray-200 {selectedFile ===
                                                        grandChild.path
                                                            ? 'bg-blue-100'
                                                            : ''}"
                                                        on:click={() =>
                                                            handleFileSelect(
                                                                grandChild.path,
                                                            )}
                                                    >
                                                        📄 {grandChild.name}
                                                    </div>
                                                {/each}
                                            </div>
                                        {/if}
                                    </div>
                                {/if}
                            {/each}
                        </div>
                    {/if}
                </div>
            {/if}
        {/each}
    </aside>

    <!-- Landing -->
    {#if !selectedFile}
        <main class="content">
            <div class="shadow document">
                <div class="heading"><h2>Mutation Testing</h2></div>
                <div class="main-content">
                    <div>
                        <span
                            >Last Ran: {new Date(
                                mutationsMeta.created_at,
                            )}</span
                        >
                    </div>
                    <div><span>For Commit: {mutationsMeta.commit}</span></div>
                    <div>Total unkilled mutants: {countMutations}</div>
                    <div>
                        <a href="{env.PUBLIC_ENDPOINT}/mutations"
                            >Raw bcore-mutation output</a
                        >
                    </div>
                    <div>
                        <br /><br />To view the mutants, select a file from the
                        left.
                    </div>
                    <div>
                        <br /><br />To see the exact bcore-mutation commands
                        that corecheck.dev ran, checkout the
                        <a
                            href="https://raw.githubusercontent.com/corecheck/corecheck/refs/heads/master/workers/mutation-worker/entrypoint.sh"
                            >entrypoint.sh</a
                        > file for the mutation worker.
                    </div>
                </div>
            </div>
        </main>
    {/if}

    <!-- File view -->
    {#if selectedFile}
        <main class="content">
            <div class="shadow document">
                <div class="heading">
                    <h2>
                        <a
                            on:click={() =>
                                goto("/mutation", { replaceState: true })}
                            style="text-decoration: underline;">Mutations</a
                        >
                        -> {selectedFile}
                    </h2>
                </div>

                <div class="main-content code-container">
                    <div class="code-scroll-wrapper">
                        {#each fileContent.split("\n") as line, index}
                            {@const lineNumber = index + 1}
                            {@const hasMutants = mutationData[lineNumber]}

                            <div>
                                <div
                                    id={"L" + lineNumber}
                                    class="line-wrapper {hasMutants
                                        ? 'red'
                                        : ''} {lineNumber === highlightedLine
                                        ? 'highlight'
                                        : ''}"
                                    on:click={() => {
                                        if (hasMutants) toggleLine(lineNumber);
                                        goto(
                                            `/mutation/${selectedFile}#L${lineNumber}`,
                                            { replaceState: true },
                                        );
                                        highlightedLine = lineNumber;
                                        // Keep current horizontal position on user click
                                        scrollLineIntoView(lineNumber, {
                                            alignLeft: false,
                                            vertical: "center",
                                        });
                                    }}
                                >
                                    <div class="lineno">{lineNumber}</div>
                                    <div class="line">
                                        {#if hasMutants}
                                            <span class="chevron"
                                                >{expandedLines.has(lineNumber)
                                                    ? "▼"
                                                    : "▶"}</span
                                            >
                                        {/if}
                                        <span><pre>{line}</pre></span>
                                    </div>
                                </div>

                                {#if expandedLines.has(lineNumber) && hasMutants}
                                    <div class="mutant-container">
                                        {#each hasMutants as mutant}
                                            <div style="margin-bottom: 1rem;">
                                                <div class="mutant-title">
                                                    Mutant #{mutant.id} - {mutant.status}
                                                </div>
                                                <div class="mutant-content">
                                                    <pre>{mutant.diff}</pre>
                                                </div>
                                            </div>
                                        {/each}
                                    </div>
                                {/if}
                            </div>
                        {/each}
                    </div>
                </div>
            </div>
        </main>
    {/if}
</div>

<style>
    /* Layout: grid + sticky sidebar */
    .page-wrapper {
        display: grid;
        grid-template-columns: 300px 1fr;
        min-height: 100vh;
    }
    .file-tree {
        position: sticky;
        top: 0;
        align-self: start;
        height: 100vh;
        overflow-y: auto;
        background: #f8f9fa;
        padding: 1rem;
        border-right: 1px solid #dee2e6;
        z-index: 1;
    }
    .content {
        padding: 1rem;
        background-color: white;
    }
    @media (max-width: 768px) {
        .page-wrapper {
            grid-template-columns: 1fr;
        }
        .file-tree {
            height: auto;
            max-height: 50vh;
        }
    }

    .directory {
        padding-left: 1.5rem;
    }
    .shadow {
        --tw-shadow:
            0 10px 15px -3px rgba(0, 0, 0, 0.1),
            0 4px 6px -2px rgba(0, 0, 0, 0.05);
        box-shadow:
            0 0 #0000,
            0 0 #0000,
            0 0 #0000,
            0 0 #0000,
            var(--tw-shadow);
    }
    .heading {
        padding: 1rem;
        border-bottom: 1px solid rgba(229, 231, 235, 1);
    }
    .document {
        max-width: 100%;
        margin: 0 auto;
        font-family:
            ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas,
            "Liberation Mono", "Courier New", monospace;
    }
    .main-content {
        padding: 1rem;
        background-color: rgba(249, 250, 251, 1);
        min-height: 100vh;
    }

    /* Code view */
    .code-container {
        overflow-x: auto;
        max-width: 100%;
        padding-left: 12px; /* left gutter */
        scroll-padding-left: 12px; /* if any future scrollIntoView is used */
    }
    .code-scroll-wrapper {
        min-width: min-content;
        white-space: nowrap;
    }
    .code-container::-webkit-scrollbar {
        height: 8px;
    }
    .code-container::-webkit-scrollbar-track {
        background: #edf2f7;
    }
    .code-container::-webkit-scrollbar-thumb {
        background-color: #cbd5e0;
        border-radius: 4px;
    }

    .line {
        color: rgba(5, 150, 105, 1);
        flex: 1 1 0%;
        display: flex;
        white-space: nowrap;
    }
    .lineno {
        color: rgba(107, 114, 128, 1);
        width: 3rem;
        flex-shrink: 0;
    }
    .line-wrapper {
        display: flex;
        align-items: flex-start;
        width: 100%;
        scroll-margin-left: 12px;
    }
    .chevron {
        margin-right: 0.5rem;
        flex-shrink: 0;
    }
    .red {
        background-color: rgba(254, 226, 226, 1);
    }

    .mutant-container {
        padding-left: 1rem;
        border-left: 2px solid rgba(229, 231, 235, 1);
        margin-left: 3rem;
        margin-bottom: 0.5rem;
        margin-top: 0.5rem;
        width: calc(100% - 3rem);
    }
    .mutant-title {
        color: rgba(75, 85, 99, 1);
    }
    .mutant-content {
        font-size: 0.875rem;
        line-height: 1.25rem;
        padding: 0.5rem;
        background-color: rgba(243, 244, 246, 1);
        overflow: hidden;
    }

    .highlight {
        outline: 2px solid rgba(59, 130, 246, 0.6);
        background-color: rgba(219, 234, 254, 0.6);
    }

    @media (max-width: 480px) {
        .lineno {
            width: 2rem;
        }
        .mutant-container {
            margin-left: 2rem;
            width: calc(100% - 2rem);
        }
    }
</style>

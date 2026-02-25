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
    let showFileTree = true;
    let isMobile = false;

    const GITHUB_RAW_BASE =
        "https://raw.githubusercontent.com/bitcoin/bitcoin/";

    let mutationsMeta = {};
    let mutations = [];
    let countMutations = 0;

    // ===== Parse git diff into lines with types =====
    function parseDiff(diffText) {
        const lines = diffText.split("\n");
        return lines.map((line) => {
            if (line.startsWith("+") && !line.startsWith("+++")) {
                return { type: "addition", text: line };
            } else if (line.startsWith("-") && !line.startsWith("---")) {
                return { type: "deletion", text: line };
            } else if (line.startsWith("@@")) {
                return { type: "hunk", text: line };
            } else {
                return { type: "context", text: line };
            }
        });
    }

    // ===== Vertical-only scroll helper =====
    function scrollLineIntoView(
        lineNumber,
        { alignLeft = false, vertical = "center", smooth = false } = {},
    ) {
        const el = document.getElementById(`L${lineNumber}`);
        if (!el) return;

        const container = document.querySelector(".code-container");

        const rect = el.getBoundingClientRect();
        const currentY = window.scrollY;
        let targetTop;

        if (vertical === "start") {
            targetTop = currentY + rect.top;
        } else if (vertical === "end") {
            targetTop = currentY + rect.bottom - window.innerHeight;
        } else {
            targetTop =
                currentY + rect.top - window.innerHeight / 2 + rect.height / 2;
        }

        window.scrollTo({
            top: targetTop,
            left: window.scrollX,
            behavior: smooth ? "smooth" : "auto",
        });

        if (container && alignLeft) {
            container.scrollLeft = 0;
        }
    }

    async function handleFileSelect(
        file,
        updateUrl = true,
        lineToFocus = null,
    ) {
        selectedFile = file;

        if (isMobile) showFileTree = false;

        if (updateUrl) {
            const currentHash =
                typeof window !== "undefined" ? window.location.hash : "";
            goto(`/mutation/${file}${currentHash}`, { replaceState: false });
        }

        try {
            const selected_mutations = mutations.filter((val) =>
                val.filename.includes(file),
            );
            mutationData =
                selected_mutations[0] && selected_mutations[0].diffs
                    ? selected_mutations[0].diffs
                    : {};

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

            await tick();

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
            "netgroup.cpp": "content",
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

    function formatDate(dateStr) {
        if (!dateStr) return "—";
        try {
            return new Date(dateStr).toLocaleString();
        } catch {
            return dateStr;
        }
    }

    onMount(async () => {
        isMobile = window.innerWidth <= 768;
        if (isMobile) showFileTree = false;

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
                    await handleFileSelect(filePath, false, targetLine);
                }
            }
        }

        const unsubscribe = page.subscribe(async ($page) => {
            const { pathname } = $page.url;
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

<div class="mutation-layout">
    <!-- Mobile-only toggle bar -->
    <div class="mobile-nav">
        <button
            class="tree-toggle-btn"
            on:click={() => (showFileTree = !showFileTree)}
            aria-label="Toggle file tree"
        >
            <span class="toggle-icon">{showFileTree ? "✕" : "☰"}</span>
            <span>{showFileTree ? "Close" : "Files"}</span>
        </button>
        {#if selectedFile}
            <span class="mobile-breadcrumb" title={selectedFile}>
                {selectedFile.split("/").pop()}
            </span>
        {:else}
            <span class="mobile-breadcrumb">Mutation Testing</span>
        {/if}
    </div>

    <!-- Sidebar: File Tree -->
    <aside class="file-tree" class:file-tree-hidden={!showFileTree}>
        <div class="tree-header">
            <h2>Files</h2>
        </div>
        <div class="tree-body">
            {#each treeData as item}
                {#if item.type === "file"}
                    <div
                        class="tree-item tree-file"
                        class:tree-selected={selectedFile === item.path}
                        on:click={() => handleFileSelect(item.path)}
                        role="button"
                        tabindex="0"
                        on:keydown={(e) => e.key === "Enter" && handleFileSelect(item.path)}
                    >
                        <span class="tree-icon">📄</span>
                        <span class="tree-name">{item.name}</span>
                    </div>
                {:else}
                    <div>
                        <div
                            class="tree-item tree-dir"
                            on:click={() => {
                                expanded = new Set(expanded);
                                if (expanded.has(item.path))
                                    expanded.delete(item.path);
                                else expanded.add(item.path);
                            }}
                            role="button"
                            tabindex="0"
                            on:keydown={(e) => {
                                if (e.key === "Enter") {
                                    expanded = new Set(expanded);
                                    if (expanded.has(item.path)) expanded.delete(item.path);
                                    else expanded.add(item.path);
                                }
                            }}
                        >
                            <span class="tree-icon">{expanded.has(item.path) ? "📂" : "📁"}</span>
                            <span class="tree-name">{item.name}</span>
                            <span class="tree-chevron">{expanded.has(item.path) ? "▾" : "▸"}</span>
                        </div>
                        {#if expanded.has(item.path)}
                            <div class="tree-children">
                                {#each item.children as child}
                                    {#if child.type === "file"}
                                        <div
                                            class="tree-item tree-file"
                                            class:tree-selected={selectedFile === child.path}
                                            on:click={() => handleFileSelect(child.path)}
                                            role="button"
                                            tabindex="0"
                                            on:keydown={(e) => e.key === "Enter" && handleFileSelect(child.path)}
                                        >
                                            <span class="tree-icon">📄</span>
                                            <span class="tree-name">{child.name}</span>
                                        </div>
                                    {:else}
                                        <div>
                                            <div
                                                class="tree-item tree-dir"
                                                on:click={() => {
                                                    expanded = new Set(expanded);
                                                    if (expanded.has(child.path))
                                                        expanded.delete(child.path);
                                                    else expanded.add(child.path);
                                                }}
                                                role="button"
                                                tabindex="0"
                                                on:keydown={(e) => {
                                                    if (e.key === "Enter") {
                                                        expanded = new Set(expanded);
                                                        if (expanded.has(child.path)) expanded.delete(child.path);
                                                        else expanded.add(child.path);
                                                    }
                                                }}
                                            >
                                                <span class="tree-icon">{expanded.has(child.path) ? "📂" : "📁"}</span>
                                                <span class="tree-name">{child.name}</span>
                                                <span class="tree-chevron">{expanded.has(child.path) ? "▾" : "▸"}</span>
                                            </div>
                                            {#if expanded.has(child.path)}
                                                <div class="tree-children">
                                                    {#each child.children as grandChild}
                                                        <div
                                                            class="tree-item tree-file"
                                                            class:tree-selected={selectedFile === grandChild.path}
                                                            on:click={() => handleFileSelect(grandChild.path)}
                                                            role="button"
                                                            tabindex="0"
                                                            on:keydown={(e) => e.key === "Enter" && handleFileSelect(grandChild.path)}
                                                        >
                                                            <span class="tree-icon">📄</span>
                                                            <span class="tree-name">{grandChild.name}</span>
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
        </div>
    </aside>

    <!-- Landing page -->
    {#if !selectedFile}
        <main class="mutation-content">
            <div class="landing-card">
                <div class="landing-header">
                    <h2>Mutation Testing</h2>
                    <p class="landing-subtitle">
                        Unkilled mutants surviving against the Bitcoin Core test suite
                    </p>
                </div>

                <div class="stat-grid">
                    <div class="stat-card stat-card-danger">
                        <div class="stat-value">{countMutations || "—"}</div>
                        <div class="stat-label">Unkilled Mutants</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value stat-mono">{mutationsMeta.commit ? mutationsMeta.commit.slice(0, 7) : "—"}</div>
                        <div class="stat-label">Commit</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value stat-date">{formatDate(mutationsMeta.created_at)}</div>
                        <div class="stat-label">Last Run</div>
                    </div>
                </div>

                <div class="landing-body">
                    <p>
                        Select a file from the sidebar to explore its surviving mutants.
                        Lines highlighted in red contain at least one unkilled mutant —
                        click a line to expand its diff.
                    </p>
                    <div class="landing-links">
                        <a
                            href="${env.PUBLIC_ENDPOINT}/mutations}"
                            target="_blank"
                            rel="noopener noreferrer"
                            class="landing-link"
                        >
                            Raw mutation output ↗
                        </a>
                        <a
                            href="https://raw.githubusercontent.com/corecheck/corecheck/refs/heads/master/workers/mutation-worker/entrypoint.sh"
                            target="_blank"
                            rel="noopener noreferrer"
                            class="landing-link"
                        >
                            Worker entrypoint.sh ↗
                        </a>
                    </div>
                </div>
            </div>
        </main>
    {/if}

    <!-- File view -->
    {#if selectedFile}
        <main class="mutation-content">
            <div class="document-card">
                <div class="doc-heading">
                    <nav class="doc-breadcrumb">
                        <button
                            class="breadcrumb-link"
                            on:click={() => goto("/mutation", { replaceState: true })}
                        >
                            Mutations
                        </button>
                        <span class="breadcrumb-sep">›</span>
                        <span class="breadcrumb-current">{selectedFile}</span>
                    </nav>
                </div>

                <div class="code-container">
                    <div class="code-scroll-wrapper">
                        {#each fileContent.split("\n") as line, index}
                            {@const lineNumber = index + 1}
                            {@const hasMutants = mutationData[lineNumber]}

                            <div>
                                <div
                                    id={"L" + lineNumber}
                                    class="line-wrapper"
                                    class:line-has-mutants={hasMutants}
                                    class:line-highlight={lineNumber === highlightedLine}
                                    on:click={() => {
                                        if (hasMutants) toggleLine(lineNumber);
                                        goto(
                                            `/mutation/${selectedFile}#L${lineNumber}`,
                                            { replaceState: true },
                                        );
                                        highlightedLine = lineNumber;
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
                                            <div class="mutant-block">
                                                <div class="mutant-title">
                                                    <span class="mutant-badge">Mutant #{mutant.id}</span>
                                                    <span class="mutant-status">{mutant.status}</span>
                                                </div>
                                                <div class="mutant-content">
                                                    {#each parseDiff(mutant.diff) as diffLine}
                                                        <div
                                                            class="diff-line diff-{diffLine.type}"
                                                        >
                                                            <pre>{diffLine.text}</pre>
                                                        </div>
                                                    {/each}
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
    /* ─── Layout ─────────────────────────────────────────────────── */
    .mutation-layout {
        display: grid;
        grid-template-columns: 280px 1fr;
        grid-template-rows: 0 1fr; /* mobile-nav row is 0 on desktop */
        min-height: 100vh;
        width: 100%;
    }

    /* ─── Mobile nav bar (hidden on desktop) ────────────────────── */
    .mobile-nav {
        display: none;
    }

    /* ─── File tree ──────────────────────────────────────────────── */
    .file-tree {
        grid-row: 1 / -1;
        position: sticky;
        top: 0;
        align-self: start;
        height: 100vh;
        overflow-y: auto;
        background: #f8f9fa;
        border-right: 1px solid #dee2e6;
        font-size: 14px;
        transition: transform 0.2s ease;
    }

    .tree-header {
        padding: 1rem 1rem 0.5rem;
        border-bottom: 1px solid #dee2e6;
        background: #f0f2f5;
        position: sticky;
        top: 0;
        z-index: 1;
    }
    .tree-header h2 {
        font-size: 13px;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.06em;
        color: #666f75;
        margin: 0;
    }
    .tree-body {
        padding: 0.5rem 0;
    }

    .tree-item {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 5px 12px;
        cursor: pointer;
        user-select: none;
        border-radius: 0;
        transition: background 0.12s;
        color: #16161a;
    }
    .tree-item:hover {
        background: #e4e9ec;
    }
    .tree-selected {
        background: #dbeafe;
        color: #1d4ed8;
        font-weight: 500;
    }
    .tree-selected:hover {
        background: #bfdbfe;
    }
    .tree-icon {
        font-size: 13px;
        flex-shrink: 0;
    }
    .tree-name {
        flex: 1;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
    .tree-chevron {
        color: #a0a6ac;
        font-size: 11px;
        flex-shrink: 0;
    }
    .tree-children {
        padding-left: 14px;
    }

    /* ─── Content area ───────────────────────────────────────────── */
    .mutation-content {
        padding: 1.25rem;
        background: #f8f9fa;
        min-width: 0;
    }

    /* ─── Landing page ───────────────────────────────────────────── */
    .landing-card {
        background: #fff;
        border-radius: 8px;
        box-shadow: 0 1px 4px rgba(0,0,0,0.08), 0 4px 16px rgba(0,0,0,0.04);
        overflow: hidden;
        max-width: 700px;
    }
    .landing-header {
        padding: 1.5rem;
        border-bottom: 1px solid #e5e7eb;
        background: #fff;
    }
    .landing-header h2 {
        font-size: 20px;
        font-weight: 600;
        color: #16161a;
        margin: 0 0 4px;
    }
    .landing-subtitle {
        color: #666f75;
        font-size: 14px;
        margin: 0;
    }

    .stat-grid {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 0;
        border-bottom: 1px solid #e5e7eb;
    }
    .stat-card {
        padding: 1.25rem 1rem;
        text-align: center;
        border-right: 1px solid #e5e7eb;
        background: #fafafa;
    }
    .stat-card:last-child {
        border-right: none;
    }
    .stat-value {
        font-size: 22px;
        font-weight: 700;
        color: #16161a;
        line-height: 1.2;
        margin-bottom: 4px;
    }
    .stat-card-danger .stat-value {
        color: #e34562;
    }
    .stat-mono {
        font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    }
    .stat-date {
        font-size: 14px;
        font-weight: 500;
    }
    .stat-label {
        font-size: 12px;
        color: #a0a6ac;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .landing-body {
        padding: 1.5rem;
        color: #444;
        font-size: 14.5px;
        line-height: 1.6;
    }
    .landing-body p {
        margin: 0 0 1rem;
    }
    .landing-links {
        display: flex;
        flex-wrap: wrap;
        gap: 0.75rem;
    }
    .landing-link {
        display: inline-flex;
        align-items: center;
        gap: 4px;
        padding: 6px 14px;
        background: #f0f2f5;
        border: 1px solid #d7dde4;
        border-radius: 6px;
        font-size: 13px;
        color: #16161a;
        text-decoration: none;
        transition: background 0.15s, border-color 0.15s;
    }
    .landing-link:hover {
        background: #e4e9ec;
        border-color: #a5b0c0;
    }

    /* ─── File / code view ───────────────────────────────────────── */
    .document-card {
        background: #fff;
        border-radius: 8px;
        box-shadow: 0 1px 4px rgba(0,0,0,0.08), 0 4px 16px rgba(0,0,0,0.04);
        overflow: hidden;
        font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas,
            "Liberation Mono", "Courier New", monospace;
    }

    .doc-heading {
        padding: 0.75rem 1rem;
        border-bottom: 1px solid #e5e7eb;
        background: #f8f9fa;
    }
    .doc-breadcrumb {
        display: flex;
        align-items: center;
        gap: 6px;
        flex-wrap: wrap;
        font-size: 14px;
    }
    .breadcrumb-link {
        background: none;
        border: none;
        padding: 0;
        cursor: pointer;
        color: #5499e8;
        font-family: inherit;
        font-size: inherit;
        text-decoration: underline;
        text-underline-offset: 2px;
    }
    .breadcrumb-link:hover {
        color: #1d4ed8;
    }
    .breadcrumb-sep {
        color: #a0a6ac;
    }
    .breadcrumb-current {
        color: #16161a;
        font-weight: 500;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        max-width: min(60vw, 500px);
    }

    .code-container {
        overflow-x: auto;
        max-width: 100%;
        padding-left: 8px;
        scroll-padding-left: 8px;
    }
    .code-container::-webkit-scrollbar { height: 6px; }
    .code-container::-webkit-scrollbar-track { background: #f0f2f5; }
    .code-container::-webkit-scrollbar-thumb {
        background: #c6cdd7;
        border-radius: 3px;
    }

    .code-scroll-wrapper {
        min-width: min-content;
        white-space: nowrap;
    }

    .line-wrapper {
        display: flex;
        align-items: flex-start;
        width: 100%;
        scroll-margin-left: 8px;
        cursor: default;
    }
    .line-wrapper:hover {
        background: #f0f2f5;
    }
    .line-has-mutants {
        background: #fee2e2;
        cursor: pointer;
    }
    .line-has-mutants:hover {
        background: #fecaca;
    }
    .line-highlight {
        outline: 2px solid rgba(59, 130, 246, 0.55);
        background: rgba(219, 234, 254, 0.7) !important;
    }

    .lineno {
        color: #a0a6ac;
        width: 3rem;
        flex-shrink: 0;
        text-align: right;
        padding-right: 12px;
        font-size: 12px;
        user-select: none;
    }
    .line {
        flex: 1 1 0%;
        display: flex;
        white-space: nowrap;
        color: #374151;
    }
    .chevron {
        margin-right: 6px;
        flex-shrink: 0;
        color: #e34562;
        font-size: 10px;
    }
    .line pre {
        margin: 0;
        font-family: inherit;
        font-size: 13px;
        line-height: 1.5;
    }

    /* ─── Mutant panel ───────────────────────────────────────────── */
    .mutant-container {
        padding: 6px 12px 6px 3.5rem;
        border-left: 3px solid #e34562;
        background: #fff5f5;
    }
    .mutant-block {
        margin-bottom: 10px;
        border: 1px solid #fecaca;
        border-radius: 6px;
        overflow: hidden;
    }
    .mutant-block:last-child {
        margin-bottom: 0;
    }
    .mutant-title {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 6px 10px;
        background: #fff0f0;
        border-bottom: 1px solid #fecaca;
        font-size: 12px;
    }
    .mutant-badge {
        font-weight: 600;
        color: #9f1239;
    }
    .mutant-status {
        color: #a0a6ac;
        font-size: 11px;
    }
    .mutant-content {
        font-size: 12px;
        line-height: 1.4;
        background: #f9fafb;
        overflow-x: auto;
    }

    .diff-line {
        margin: 0;
        padding: 2px 10px;
        font-family: ui-monospace, Menlo, "Courier New", monospace;
    }
    .diff-line pre {
        margin: 0;
        padding: 0;
        white-space: pre;
        font-family: inherit;
        font-size: 12px;
    }
    .diff-addition {
        background: #d4edda;
        color: #155724;
    }
    .diff-deletion {
        background: #f8d7da;
        color: #721c24;
    }
    .diff-hunk {
        background: #e7f3ff;
        color: #004085;
        font-weight: 600;
    }
    .diff-context {
        background: transparent;
        color: #374151;
    }

    /* ─── Mobile (≤ 768px) ───────────────────────────────────────── */
    @media (max-width: 768px) {
        .mutation-layout {
            grid-template-columns: 1fr;
            grid-template-rows: auto auto 1fr;
        }

        .mobile-nav {
            display: flex;
            align-items: center;
            gap: 10px;
            padding: 8px 12px;
            background: #fff;
            border-bottom: 1px solid #dee2e6;
            position: sticky;
            top: 0;
            z-index: 10;
        }

        .tree-toggle-btn {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            padding: 6px 12px;
            background: #f0f2f5;
            border: 1px solid #d7dde4;
            border-radius: 6px;
            font-size: 13px;
            font-weight: 500;
            color: #16161a;
            cursor: pointer;
            flex-shrink: 0;
            transition: background 0.15s;
        }
        .tree-toggle-btn:hover {
            background: #e4e9ec;
        }
        .toggle-icon {
            font-style: normal;
        }

        .mobile-breadcrumb {
            font-size: 13px;
            color: #666f75;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            flex: 1;
        }

        .file-tree {
            position: static;
            height: auto;
            max-height: 55vh;
            width: 100%;
            border-right: none;
            border-bottom: 1px solid #dee2e6;
        }
        .file-tree-hidden {
            display: none;
        }

        .mutation-content {
            padding: 0.75rem;
        }

        .stat-grid {
            grid-template-columns: 1fr;
        }
        .stat-card {
            border-right: none;
            border-bottom: 1px solid #e5e7eb;
            padding: 1rem;
        }
        .stat-card:last-child {
            border-bottom: none;
        }
        .landing-card {
            max-width: 100%;
        }

        .lineno {
            width: 2.5rem;
            padding-right: 8px;
        }
        .mutant-container {
            padding-left: 2.5rem;
        }
        .breadcrumb-current {
            max-width: 45vw;
        }
    }

    @media (max-width: 480px) {
        .lineno {
            width: 2rem;
            font-size: 11px;
        }
        .line pre {
            font-size: 12px;
        }
    }
</style>

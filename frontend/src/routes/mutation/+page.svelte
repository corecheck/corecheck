<script>
  let selectedFile = '';
  let fileContent = '';
  let mutationData = {};
  let expandedLines = new Set();
  let expanded = new Set(['src', 'src/script', 'src/wallet']);

  const GITHUB_RAW_BASE = 'https://raw.githubusercontent.com/bitcoin/bitcoin/master';
  const files = {
    'src': {
      'wallet': {
        'coinselection.cpp': 'content',
      },
      'script': {
        'interpreter.cpp': 'content'
      },
      'consensus': {
        'tx_verify.cpp': 'content',
        'tx_check.cpp': 'content',
        'merkle.cpp': 'content'
      },
      'util': {
        'asmap.cpp': 'content'
      },
      'pow.cpp': 'content',
      'addrman.cpp': 'content'
    },
  };

  async function handleFileSelect(file) {
    selectedFile = file;
    try {
      // Fetch mutations from local public directory
      const mutationsResp = await fetch('https://api-dev.corecheck.dev/mutations');
      const mutations = await mutationsResp.json();
      const selected_mutations = mutations.filter(val => val.filename.includes(file));

      if(selected_mutations.length > 0 && 'diffs' in selected_mutations[0]) {
        mutationData = selected_mutations[0].diffs || {};
      }

      // Fetch file content from GitHub raw URL
      const githubPath = `${GITHUB_RAW_BASE}/${file}`;
      const contentResp = await fetch(githubPath);
      if (!contentResp.ok) {
        throw new Error(`Failed to fetch file: ${contentResp.status}`);
      }
      fileContent = await contentResp.text();
    } catch (error) {
      console.error('Error:', error);
      fileContent = `Error loading content: ${error.message}`;
      mutationData = {};
    }
  }

  function toggleLine(lineNumber) {
    expandedLines = new Set(expandedLines);
    if (expandedLines.has(lineNumber)) {
      expandedLines.delete(lineNumber);
    } else {
      expandedLines.add(lineNumber);
    }
  }

  function toggleDir(path) {
    expanded = new Set(expanded);
    if (expanded.has(path)) {
      expanded.delete(path);
    } else {
      expanded.add(path);
    }
  }

  function renderTree(tree, path = '') {
    return Object.entries(tree).map(([name, value]) => {
      const currentPath = path ? `${path}/${name}` : name;
      if (typeof value === 'string') {
        return {
          type: 'file',
          name,
          path: currentPath
        };
      } else {
        return {
          type: 'directory',
          name,
          path: currentPath,
          children: renderTree(value, currentPath)
        };
      }
    });
  }

  $: treeData = renderTree(files);
</script>

<div class="page-wrapper">
  <!-- File Tree -->
  <div class="file-tree">
    <h2 class="text-xl font-bold mb-4">Files</h2>
    {#each treeData as item}
      {#if item.type === 'file'}
        <div
          class="pl-6 py-1 cursor-pointer hover:bg-gray-200 {selectedFile === item.path ? 'bg-blue-100' : ''}"
          on:click={() => handleFileSelect(item.path)}
        >
          📄 {item.name}
        </div>
      {:else}
        <div>
          <div
            class="py-1 cursor-pointer hover:bg-gray-200 flex items-center"
            on:click={() => toggleDir(item.path)}
          >
            <span class="mr-2">{expanded.has(item.path) ? '📂' : '📁'}</span>
            {item.name}
          </div>
          {#if expanded.has(item.path)}
            <div class="directory">
              {#each item.children as child}
                {#if child.type === 'file'}
                  <div
                    class="pl-6 py-1 cursor-pointer hover:bg-gray-200 {selectedFile === child.path ? 'bg-blue-100' : ''}"
                    on:click={() => handleFileSelect(child.path)}
                  >
                    📄 {child.name}
                  </div>
                {:else}
                  <div>
                    <div
                      class="py-1 cursor-pointer hover:bg-gray-200 flex items-center"
                      on:click={() => toggleDir(child.path)}
                    >
                      <span class="mr-2">{expanded.has(child.path) ? '📂' : '📁'}</span>
                      {child.name}
                    </div>
                    {#if expanded.has(child.path)}
                      <div class="directory">
                        {#each child.children as grandChild}
                          <div
                            class="pl-6 py-1 cursor-pointer hover:bg-gray-200 {selectedFile === grandChild.path ? 'bg-blue-100' : ''}"
                            on:click={() => handleFileSelect(grandChild.path)}
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
  </div>

  {#if !selectedFile}
    <div class="content">
      <div class="">
        <div class="shadow document" style="">
          <div class="heading" style="">
            <h2 class="">Mutation Testing</h2>
          </div>

          <div class="main-content">
            <div>
              <span>Last Ran: 2025-02-21 12:05</span>
            </div>
            <div>
              <span>For Commit: 879569cab4e5b400350f3b95d7bee71b49636591</span>
            </div>
            <div>
              Total unkilled mutants: 34
            </div>
            <div>
              <a href="https://api-dev.corecheck.dev/mutations">Raw mutation-core output</a>
            </div>
            <div>
              <br><br>
              To view the mutants, select a file from the left.
            </div>
            <div>
              <br><br>
              To see the exact mutation-core commands that corecheck.dev ran, checkout the <a href="https://raw.githubusercontent.com/corecheck/corecheck/refs/heads/master/workers/mutation-worker/entrypoint.sh">entrypoint.sh</a> file for the mutation worker.
            </div>
          </div>
        </div>
      </div>
    </div>


  {/if}
  <!-- Content View -->
  {#if selectedFile}
    <div class="content">
      <div class="">
        <div class="shadow document" style="">
          <div class="heading" style="">
            <h2 class="">{selectedFile}</h2>
          </div>

          <div class="main-content">
              {#each fileContent.split('\n') as line, index}
                {@const lineNumber = index + 1}
                {@const hasMutants = mutationData[lineNumber]}
                {@const lineColor = hasMutants ? 'text-red-600' : 'text-green-600'}

                <div class="">
                  <div
                    style=""
                    class="line-wrapper {hasMutants ? 'red' : ''}"
                    on:click={() => hasMutants && toggleLine(lineNumber)}
                  >
                    <div class="lineno" style="">
                      {lineNumber}
                    </div>
                    <div class="line">
                      {#if hasMutants}
                        <span class="chevron">
                          {expandedLines.has(lineNumber) ? '▼' : '▶'}
                        </span>
                      {/if}
                      <span><pre>{line}</pre></span>
                    </div>
                  </div>

                  {#if expandedLines.has(lineNumber) && hasMutants}
                    <div class="mutant-container">
                      {#each hasMutants as mutant}
                        <div class="" style="margin-bottom: 1rem;">
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
    </div>
  {/if}
</div>

<style>
  .file-tree {
    position: fixed;
    width: 300px;
    height: 100vh;
    overflow-y: auto;
    background: #f8f9fa;
    padding: 1rem;
    border-right: 1px solid #dee2e6;
  }

  .content {
    margin-left: 300px;
    padding: 1rem;
    background-color: white;
  }
  .content a {
    cursor: pointer;
  }

  .directory {
    padding-left: 1.5rem;
  }
  .shadow {
    --tw-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1),0 4px 6px -2px rgba(0, 0, 0, 0.05);
    box-shadow: 0 0 #0000,0 0 #0000, 0 0 #0000,0 0 #0000,var(--tw-shadow);
  }
  .heading {
    padding: 1rem;
    border-bottom-width: 1px;
    border-style: solid;
    border-color: rgba(229, 231, 235, 1);
  }
  .document {
    max-width: 72rem;
    margin-right: auto;
    margin-left: 50px;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  }
  .main-content {
    padding: 1rem;
    background-color: rgba(249,250,251,1);
    min-height: 100vh;
  }
  .line {
    color: rgba(5,150,105,1);
    flex: 1 1 0%;
    overflow: hidden;
    display: flex;
  }
  .lineno {
    color: rgba(107,114,128,1);
    width: 3rem;
  }
  .line-wrapper {
    display: flex;
    align-items: flex-start;
  }
  .chevron {
    margin-right: 0.5rem;
  }
  .red {
    background-color: rgba(254,226,226,1);
  }
  .mutant-container {
    padding-left: 1rem;
    border-style: solid;
    border-left-width: 2px;
    border-color: rgba(229,231,235,1);
    margin-left: 3rem;
    margin-bottom: .5rem;
    margin-top: .5rem;
  }
  .mutant-title {
    color: rgba(75,85,99,1);
  }
  .mutant-content {
    font-size: .875rem;
    line-height: 1.25rem;
    padding: 0.5rem;
    background-color: rgba(243,244,246,1);
    overflow: hidden;
  }
</style>

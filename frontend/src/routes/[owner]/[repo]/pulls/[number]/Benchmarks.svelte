<script lang="ts">
    import tooltip from "@/actions/tooltip";
    import Accordion from "@/components/base/Accordion.svelte";
    import Field from "@/components/base/Field.svelte";
    import SortHeader from "@/components/base/SortHeader.svelte";
    import Toggler from "@/components/base/Toggler.svelte";
    let sort = "-diff";

    export let report: any;
    let showOnlySignificant = true;
    const threshold = 0.10;

    function displayBenchNumber(n, showSign = false) {
        if (!n) return 0;
        return Math.round(n).toLocaleString("en-US", {
            minimumFractionDigits: 0,
            maximumFractionDigits: 0,
        });
    }

    function displayPercentage(n) {
        const r = (n * 100).toLocaleString("en-US", {
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
        });

        if (n > 0) return "+" + r;
        return r;
    }

    function getUnit(benchmark) {
        if (!report.benchmarks_grouped[benchmark]) return "";
        return report.benchmarks_grouped[benchmark]["unit"];
    }

    function getNsPerUnit(benchmark) {
        if (!report.benchmarks_grouped[benchmark]) return 0;
        return (
            report.benchmarks_grouped[benchmark]["median(elapsed)"] /
            (0.000000001 * report.benchmarks_grouped[benchmark]["batch"])
        );
    }

    function getNsPerUnitMaster(benchmark) {
        if (!report.base_report.benchmarks_grouped[benchmark]) return 0;
        return (
            report.base_report.benchmarks_grouped[benchmark][
                "median(elapsed)"
            ] /
            (0.000000001 *
                report.base_report.benchmarks_grouped[benchmark]["batch"])
        );
    }

    function getNsPerUnitDiff(benchmark) {
        if (!report.base_report.benchmarks_grouped[benchmark]) return 0;
        return (
            (getNsPerUnit(benchmark) - getNsPerUnitMaster(benchmark)) /
            getNsPerUnitMaster(benchmark)
        );
    }

    function getUnitPerSecond(benchmark) {
        if (!report.benchmarks_grouped[benchmark]) return 0;
        return (
            report.benchmarks_grouped[benchmark]["batch"] /
            report.benchmarks_grouped[benchmark]["median(elapsed)"]
        );
    }

    function getUnitPerSecondMaster(benchmark) {
        if (!report.base_report.benchmarks_grouped[benchmark]) return 0;
        return (
            report.base_report.benchmarks_grouped[benchmark]["batch"] /
            report.base_report.benchmarks_grouped[benchmark]["median(elapsed)"]
        );
    }

    function getUnitPerSecondDiff(benchmark) {
        if (!report.base_report.benchmarks_grouped[benchmark]) return 0;
        return (
            (getUnitPerSecond(benchmark) - getUnitPerSecondMaster(benchmark)) /
            getUnitPerSecondMaster(benchmark)
        );
    }

    function getIPC(benchmark) {
        if (!report.benchmarks_grouped[benchmark]) return 0;
        return (
            report.benchmarks_grouped[benchmark]["median(instructions)"] /
            report.benchmarks_grouped[benchmark]["median(cpucycles)"]
        );
    }

    function getIPCMaster(benchmark) {
        if (!report.base_report.benchmarks_grouped[benchmark]) return 0;
        return (
            report.base_report.benchmarks_grouped[benchmark][
                "median(instructions)"
            ] /
            report.base_report.benchmarks_grouped[benchmark][
                "median(cpucycles)"
            ]
        );
    }

    function getIPCDiff(benchmark) {
        if (!report.base_report.benchmarks_grouped[benchmark]) return 0;
        return (
            (getIPC(benchmark) - getIPCMaster(benchmark)) /
            getIPCMaster(benchmark)
        );
    }

    function getCyclesPerUnit(benchmark) {
        if (!report.benchmarks_grouped[benchmark]) return 0;
        return (
            report.benchmarks_grouped[benchmark]["median(cpucycles)"] /
            report.benchmarks_grouped[benchmark]["batch"]
        );
    }

    function getCyclesPerUnitMaster(benchmark) {
        if (!report.base_report.benchmarks_grouped[benchmark]) return 0;
        return (
            report.base_report.benchmarks_grouped[benchmark][
                "median(cpucycles)"
            ] / report.base_report.benchmarks_grouped[benchmark]["batch"]
        );
    }

    function getCyclesPerUnitDiff(benchmark) {
        if (!report.base_report.benchmarks_grouped[benchmark]) return 0;
        return (
            (getCyclesPerUnit(benchmark) - getCyclesPerUnitMaster(benchmark)) /
            getCyclesPerUnitMaster(benchmark)
        );
    }

    function getInstructionsPerUnit(benchmark) {
        if (!report.benchmarks_grouped[benchmark]) return 0;
        return (
            report.benchmarks_grouped[benchmark]["median(instructions)"] /
            report.benchmarks_grouped[benchmark]["batch"]
        );
    }

    function getInstructionsPerUnitMaster(benchmark) {
        if (!report.base_report.benchmarks_grouped[benchmark]) return 0;
        return (
            report.base_report.benchmarks_grouped[benchmark][
                "median(instructions)"
            ] / report.base_report.benchmarks_grouped[benchmark]["batch"]
        );
    }

    function getInstructionsPerUnitDiff(benchmark) {
        if (!report.base_report.benchmarks_grouped[benchmark]) return 0;
        return (
            (getInstructionsPerUnit(benchmark) -
                getInstructionsPerUnitMaster(benchmark)) /
            getInstructionsPerUnitMaster(benchmark)
        );
    }

    function getBranchesPerUnit(benchmark) {
        if (!report.benchmarks_grouped[benchmark]) return 0;
        return (
            report.benchmarks_grouped[benchmark]["median(branchinstructions)"] /
            report.benchmarks_grouped[benchmark]["batch"]
        );
    }

    function getBranchesPerUnitMaster(benchmark) {
        if (!report.base_report.benchmarks_grouped[benchmark]) return 0;
        return (
            report.base_report.benchmarks_grouped[benchmark][
                "median(branchinstructions)"
            ] / report.base_report.benchmarks_grouped[benchmark]["batch"]
        );
    }

    function getBranchesPerUnitDiff(benchmark) {
        return (
            (getBranchesPerUnit(benchmark) -
                getBranchesPerUnitMaster(benchmark)) /
            getBranchesPerUnitMaster(benchmark)
        );
    }

    function isSignificant(benchmark) {
        return (
            Math.abs(getNsPerUnitDiff(benchmark)) > threshold ||
            Math.abs(getUnitPerSecondDiff(benchmark)) > threshold ||
            Math.abs(getInstructionsPerUnitDiff(benchmark)) > threshold ||
            Math.abs(getCyclesPerUnitDiff(benchmark)) > threshold ||
            Math.abs(getIPCDiff(benchmark)) > threshold ||
            Math.abs(getBranchesPerUnitDiff(benchmark)) > threshold
        );
    }
</script>

<div class="flex">
    <h1>
        Benchmarks <span class="label label-success">Beta</span>
    </h1>
</div>
<div class="clearfix m-b-base" />
{#if report.benchmark_status === "pending"}
    <div class="alert alert-warning" style="text-align: center">
        <i class="ri-information-line" /> Benchmarks are currently being generated,
        please come back later.
    </div>
{:else if report.benchmark_status === "failure"}
    <div class="alert alert-danger" style="text-align: center">
        <i class="ri-information-line" /> An error occured while generating the benchmarks. Push a new commit to re-run the benchmarks.
    </div>
{:else if report.benchmark_status === "not_found"}
    <div class="alert alert-info" style="text-align: center">
        <i class="ri-information-line" /> No benchmarks available. Push a new commit to re-run the benchmarks.
    </div>
{:else if report.benchmark_status === "success"}
    <Field class="form-field form-field-toggle" name="verified" let:uniqueId>
        <input
            type="checkbox"
            id={uniqueId}
            bind:checked={showOnlySignificant}
        />
        <label for={uniqueId}
            >Only show benchmarks with a significant difference</label
        >
    </Field>
    <table class="table">
        <thead>
            <tr>
                <SortHeader name="name" bind:sort>
                    <div class="col-header-content">
                        <i class="ri-text" />
                        <span class="txt">Name</span>
                    </div>
                </SortHeader>
                <SortHeader
                    class="col-type-number col-field-type"
                    name="ns/unit"
                    bind:sort
                >
                    <div class="col-header-content">
                        <i class="ri-percent-line" />
                        <span class="txt">ns/unit</span>
                    </div>
                </SortHeader>
                <SortHeader
                    class="col-type-number col-field-type"
                    name="unit/s"
                    bind:sort
                >
                    <div class="col-header-content">
                        <i class="ri-percent-line" />
                        <span class="txt">unit/s</span>
                    </div>
                </SortHeader>
                <SortHeader
                    class="col-type-number col-field-type"
                    name="ins/unit"
                    bind:sort
                >
                    <div class="col-header-content">
                        <i class="ri-cpu-line" />
                        <span class="txt">ins/unit</span>
                    </div>
                </SortHeader>
                <SortHeader
                    class="col-type-number col-field-type"
                    name="cyc/unit"
                    bind:sort
                >
                    <div class="col-header-content">
                        <i class="ri-database-2-line" />
                        <span class="txt">cyc/unit</span>
                    </div>
                </SortHeader>
                <SortHeader
                    class="col-type-number col-field-type"
                    name="ipc"
                    bind:sort
                >
                    <div class="col-header-content">
                        <i class="ri-database-2-line" />
                        <span class="txt">IPC</span>
                    </div>
                </SortHeader>
                <SortHeader
                    class="col-type-number col-field-type"
                    name="totalTime"
                    bind:sort
                >
                    <div class="col-header-content">
                        <i class="ri-time-line" />
                        <span class="txt">Total time</span>
                    </div>
                </SortHeader>
            </tr></thead
        >

        <tbody>
            {#each Object.keys(report.benchmarks_grouped)
                .sort((a, b) => {
                    const benchA = report.benchmarks_grouped[a];
                    const benchB = report.benchmarks_grouped[b];

                    if (sort === "+name") return benchA.name.localeCompare(benchB.name);
                    if (sort === "-name") return benchB.name.localeCompare(benchA.name);

                    if (sort === "+ns/unit") return getNsPerUnit(a) - getNsPerUnit(b);
                    if (sort === "-ns/unit") return getNsPerUnit(b) - getNsPerUnit(a);

                    if (sort === "+unit/s") return getUnitPerSecond(a) - getUnitPerSecond(b);
                    if (sort === "-unit/s") return getUnitPerSecond(b) - getUnitPerSecond(a);

                    if (sort === "+ins/unit") return getInstructionsPerUnit(a) - getInstructionsPerUnit(b);
                    if (sort === "-ins/unit") return getInstructionsPerUnit(b) - getInstructionsPerUnit(a);

                    if (sort === "+cyc/unit") return getCyclesPerUnit(a) - getCyclesPerUnit(b);
                    if (sort === "-cyc/unit") return getCyclesPerUnit(b) - getCyclesPerUnit(a);

                    if (sort === "+ipc") return getIPC(a) - getIPC(b);
                    if (sort === "-ipc") return getIPC(b) - getIPC(a);

                    if (sort === "+bra/unit") return getBranchesPerUnit(a) - getBranchesPerUnit(b);
                    if (sort === "-bra/unit") return getBranchesPerUnit(b) - getBranchesPerUnit(a);

                    return 0;
                })
                .filter((b) => !["AddrManSelectFromAlmostEmpty", "RollingBloomReset", "AddrManGetAddr", "LoadExternalBlockFile", "PrevectorDeserializeNontrivial", "GCSFilterConstruct"].includes(b))
                .filter((b) => !showOnlySignificant || isSignificant(b)) as benchmark}
                <tr>
                    <td class="col-field-id">
                        <p
                            use:tooltip={report.benchmarks_grouped[benchmark]
                                .name.length > 50
                                ? report.benchmarks_grouped[benchmark].name
                                : ""}
                        >
                            {report.benchmarks_grouped[benchmark].name.length >
                            50
                                ? report.benchmarks_grouped[
                                      benchmark
                                  ].name.substring(0, 50) + "..."
                                : report.benchmarks_grouped[benchmark].name}
                        </p>
                    </td>
                    <td class="col-type-number col-field-pr">
                        <p
                            use:tooltip={{
                                text: `master: ${displayBenchNumber(
                                    getNsPerUnitMaster(benchmark),
                                )}`,
                                position: "left",
                            }}
                        >
                            {displayBenchNumber(getNsPerUnit(benchmark))}
                            <small class="txt-hint"
                                >ns/{getUnit(benchmark)}</small
                            >
                            <small
                                class:txt-success={getNsPerUnitDiff(benchmark) <
                                    -threshold}
                                class:txt-danger={getNsPerUnitDiff(benchmark) >
                                    threshold}
                                class:txt-hint={getNsPerUnitDiff(benchmark) >=
                                    -threshold &&
                                    getNsPerUnitDiff(benchmark) <= threshold}
                            >
                                {displayPercentage(
                                    getNsPerUnitDiff(benchmark),
                                )}%
                            </small>
                        </p>
                    </td>
                    <td class="col-type-number col-field-pr">
                        <p
                            use:tooltip={{
                                text: `master: ${displayBenchNumber(
                                    getUnitPerSecondMaster(benchmark),
                                )}`,
                                position: "left",
                            }}
                        >
                            {displayBenchNumber(getUnitPerSecond(benchmark))}
                            <small class="txt-hint"
                                >{getUnit(benchmark)}/s</small
                            >
                            <small
                                class:txt-success={getUnitPerSecondDiff(
                                    benchmark,
                                ) > threshold}
                                class:txt-danger={getUnitPerSecondDiff(
                                    benchmark,
                                ) < -threshold}
                                class:txt-hint={getUnitPerSecondDiff(
                                    benchmark,
                                ) <= threshold &&
                                    getUnitPerSecondDiff(benchmark) >= -threshold}
                            >
                                {displayPercentage(
                                    getUnitPerSecondDiff(benchmark),
                                )}%
                            </small>
                        </p>
                    </td>
                    <td class="col-type-number col-field-pr">
                        <p
                            use:tooltip={{
                                text: `master: ${displayBenchNumber(
                                    getInstructionsPerUnitMaster(benchmark),
                                )}`,
                                position: "left",
                            }}
                        >
                            {displayBenchNumber(
                                getInstructionsPerUnit(benchmark),
                            )}
                            <small class="txt-hint"
                                >ins/{getUnit(benchmark)}</small
                            >
                            <small
                                class:txt-success={getInstructionsPerUnitDiff(
                                    benchmark,
                                ) < -threshold}
                                class:txt-danger={getInstructionsPerUnitDiff(
                                    benchmark,
                                ) > threshold}
                                class:txt-hint={getInstructionsPerUnitDiff(
                                    benchmark,
                                ) >= -threshold &&
                                    getInstructionsPerUnitDiff(benchmark) <=
                                        threshold}
                            >
                                {displayPercentage(
                                    getInstructionsPerUnitDiff(benchmark),
                                )}%
                            </small>
                        </p>
                    </td>
                    <td class="col-type-number col-field-pr">
                        <p
                            use:tooltip={{
                                text: `master: ${displayBenchNumber(
                                    getCyclesPerUnitMaster(benchmark),
                                )}`,
                                position: "left",
                            }}
                        >
                            {displayBenchNumber(getCyclesPerUnit(benchmark))}
                            <small class="txt-hint"
                                >cyc/{getUnit(benchmark)}</small
                            >
                            <small
                                class:txt-success={getCyclesPerUnitDiff(
                                    benchmark,
                                ) < -threshold}
                                class:txt-danger={getCyclesPerUnitDiff(
                                    benchmark,
                                ) > threshold}
                                class:txt-hint={getCyclesPerUnitDiff(
                                    benchmark,
                                ) >= -threshold &&
                                    getCyclesPerUnitDiff(benchmark) <= threshold}
                            >
                                {displayPercentage(
                                    getCyclesPerUnitDiff(benchmark),
                                )}%
                            </small>
                        </p>
                    </td>
                    <td class="col-type-number col-field-pr">
                        <p
                            use:tooltip={{
                                text: `master: ${getIPCMaster(
                                    benchmark,
                                ).toLocaleString("en-US", {
                                    minimumFractionDigits: 2,
                                    maximumFractionDigits: 2,
                                })}`,
                                position: "left",
                            }}
                        >
                            {getIPC(benchmark).toLocaleString("en-US", {
                                minimumFractionDigits: 2,
                                maximumFractionDigits: 2,
                            })}
                            <small class="txt-hint">IPC</small>
                            <small
                                class:txt-danger={getIPCDiff(benchmark) < -threshold}
                                class:txt-success={getIPCDiff(benchmark) > threshold}
                                class:txt-hint={getIPCDiff(benchmark) >=
                                    -threshold && getIPCDiff(benchmark) <= threshold}
                            >
                                {displayPercentage(getIPCDiff(benchmark))}%
                            </small>
                        </p>
                    </td>
                    <td class="col-type-number col-field-pr">
                        <p
                            use:tooltip={{
                                text: report.base_report.benchmarks_grouped[benchmark]
                                    ? `master: ${report.base_report.benchmarks_grouped[
                                    benchmark
                                ]["totalTime"].toLocaleString("en-US", {
                                    minimumFractionDigits: 2,
                                    maximumFractionDigits: 2,
                                })}`
                                : "",
                                position: "left",
                            }}
                        >
                            {report.benchmarks_grouped[benchmark][
                                "totalTime"
                            ].toLocaleString("en-US", {
                                minimumFractionDigits: 2,
                                maximumFractionDigits: 2,
                            })}
                            <small class="txt-hint">seconds</small>
                        </p>
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>
{/if}

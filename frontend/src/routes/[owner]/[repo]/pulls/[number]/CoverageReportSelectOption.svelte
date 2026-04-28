<script lang="ts">
    import dayjs from "dayjs";
    import relativeTime from "dayjs/plugin/relativeTime";
    dayjs.extend(relativeTime);

    type CoverageReportOption = {
        commit: string;
        created_at: string;
        status?: string;
    };

    export let item: CoverageReportOption;

    function getStatusLabel(status?: string) {
        if (status === "pending") return "PENDING";
        if (status === "failure") return "FAILED";
        return "";
    }
</script>

<span class="dropdown-item-text"
    >{item.commit.substring(0, 7)} - {dayjs().to(dayjs(item.created_at))} {getStatusLabel(
        item.status,
    )
        ? `(${getStatusLabel(item.status)})`
        : ""}</span
>

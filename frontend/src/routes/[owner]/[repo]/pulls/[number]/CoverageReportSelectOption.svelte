<script>
    import dayjs from "dayjs";
    import relativeTime from "dayjs/plugin/relativeTime";
    dayjs.extend(relativeTime);

    export let item;
</script>

<span class="dropdown-item-text">{item.commit.substring(0, 7)} - {dayjs().to(dayjs(item.created_at))} {item.status === "pending" ? "(PENDING)" : ""}</span>
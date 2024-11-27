import { c as create_ssr_component, b as createEventDispatcher, a as escape, d as add_attribute, v as validate_component, e as each } from "../../../../../chunks/ssr.js";
import { C as CommonHelper, S as SortHeader } from "../../../../../chunks/SortHeader.js";
const RefreshButton_svelte_svelte_type_style_lang = "";
const css$2 = {
  code: "@keyframes svelte-1bvelc2-refresh{100%{transform:rotate(180deg)}}.btn.refreshing.svelte-1bvelc2 i.svelte-1bvelc2{animation:svelte-1bvelc2-refresh 150ms ease-out}",
  map: null
};
const RefreshButton = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  createEventDispatcher();
  let { tooltip: tooltipData = { text: "Refresh", position: "right" } } = $$props;
  let { class: classes = "" } = $$props;
  if ($$props.tooltip === void 0 && $$bindings.tooltip && tooltipData !== void 0)
    $$bindings.tooltip(tooltipData);
  if ($$props.class === void 0 && $$bindings.class && classes !== void 0)
    $$bindings.class(classes);
  $$result.css.add(css$2);
  return `<button type="button" aria-label="Refresh" class="${[
    "btn btn-transparent btn-circle " + escape(classes, true) + " svelte-1bvelc2",
    ""
  ].join(" ").trim()}"><i class="ri-refresh-line svelte-1bvelc2"></i> </button>`;
});
const Searchbar = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  createEventDispatcher();
  const uniqueId = "search_" + CommonHelper.randomString(7);
  let { value = "" } = $$props;
  let { placeholder = 'Search term or filter like created > "2022-01-01"...' } = $$props;
  let searchInput;
  let tempValue = "";
  if ($$props.value === void 0 && $$bindings.value && value !== void 0)
    $$bindings.value(value);
  if ($$props.placeholder === void 0 && $$bindings.placeholder && placeholder !== void 0)
    $$bindings.placeholder(placeholder);
  {
    if (typeof value === "string") {
      tempValue = value;
    }
  }
  return ` <form class="searchbar"><label${add_attribute("for", uniqueId, 0)} class="m-l-10 txt-xl"><i class="ri-search-line"></i></label> <input type="text"${add_attribute("id", uniqueId, 0)}${add_attribute("placeholder", value || placeholder, 0)}${add_attribute("this", searchInput, 0)}${add_attribute("value", tempValue, 0)}> ${(value.length || tempValue.length) && tempValue != value ? `<button type="submit" class="btn btn-expanded btn-sm btn-warning" data-svelte-h="svelte-x7m777"><span class="txt">Search</span></button>` : ``} ${value.length || tempValue.length ? `<button type="button" class="btn btn-transparent btn-sm btn-hint p-l-xs p-r-xs m-l-10" data-svelte-h="svelte-1j0eomg"><span class="txt">Clear</span></button>` : ``}</form>`;
});
const FormattedDate_svelte_svelte_type_style_lang = "";
const css$1 = {
  code: ".datetime.svelte-zdiknu{width:100%;display:block;line-height:var(--smLineHeight)}.time.svelte-zdiknu{font-size:var(--smFontSize);color:var(--txtHintColor)}",
  map: null
};
const FormattedDate = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let dateOnly;
  let timeOnly;
  let { date = "" } = $$props;
  if ($$props.date === void 0 && $$bindings.date && date !== void 0)
    $$bindings.date(date);
  $$result.css.add(css$1);
  dateOnly = date ? date.substring(0, 10) : null;
  timeOnly = date ? date.substring(10, 19) : null;
  return `${date ? `<div class="datetime svelte-zdiknu"><div class="date">${escape(dateOnly)}</div> <div class="time svelte-zdiknu">${escape(timeOnly)} UTC</div></div>` : `<span class="txt txt-hint" data-svelte-h="svelte-1qnavrh">N/A</span>`}`;
});
const HorizontalScroller_svelte_svelte_type_style_lang = "";
const css = {
  code: ".horizontal-scroller.svelte-wc2j9h{width:auto;overflow-x:auto}.horizontal-scroller-wrapper.svelte-wc2j9h{position:relative}.horizontal-scroller-wrapper .columns-dropdown{top:40px;z-index:100;max-height:340px}",
  map: null
};
const HorizontalScroller = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let { class: classes = "" } = $$props;
  let wrapper = null;
  let scrollClasses = "";
  function refresh() {
    {
      return;
    }
  }
  if ($$props.class === void 0 && $$bindings.class && classes !== void 0)
    $$bindings.class(classes);
  if ($$props.refresh === void 0 && $$bindings.refresh && refresh !== void 0)
    $$bindings.refresh(refresh);
  $$result.css.add(css);
  return ` <div class="horizontal-scroller-wrapper svelte-wc2j9h">${slots.before ? slots.before({}) : ``} <div class="${"horizontal-scroller " + escape(classes, true) + " " + escape(scrollClasses, true) + " svelte-wc2j9h"}"${add_attribute("this", wrapper, 0)}>${slots.default ? slots.default({}) : ``}</div> ${slots.after ? slots.after({}) : ``} </div>`;
});
const PRList = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let { filter = "" } = $$props;
  let { presets = "" } = $$props;
  let { sort = "-rowid" } = $$props;
  let { items = [] } = $$props;
  if ($$props.filter === void 0 && $$bindings.filter && filter !== void 0)
    $$bindings.filter(filter);
  if ($$props.presets === void 0 && $$bindings.presets && presets !== void 0)
    $$bindings.presets(presets);
  if ($$props.sort === void 0 && $$bindings.sort && sort !== void 0)
    $$bindings.sort(sort);
  if ($$props.items === void 0 && $$bindings.items && items !== void 0)
    $$bindings.items(items);
  let $$settled;
  let $$rendered;
  do {
    $$settled = true;
    $$rendered = `${validate_component(HorizontalScroller, "HorizontalScroller").$$render($$result, { class: "table-wrapper" }, {}, {
      default: () => {
        return `<table class="${["table", ""].join(" ").trim()}"><thead><tr>${validate_component(SortHeader, "SortHeader").$$render(
          $$result,
          {
            disable: true,
            class: "col-type-number col-field-status",
            name: "status",
            sort
          },
          {
            sort: ($$value) => {
              sort = $$value;
              $$settled = false;
            }
          },
          {
            default: () => {
              return `<div class="col-header-content" data-svelte-h="svelte-qec44j"><i${add_attribute("class", CommonHelper.getFieldTypeIcon("status"), 0)}></i> <span class="txt">Status</span></div>`;
            }
          }
        )} ${validate_component(SortHeader, "SortHeader").$$render(
          $$result,
          {
            disable: true,
            class: "col-type-number col-field-number",
            name: "number",
            sort
          },
          {
            sort: ($$value) => {
              sort = $$value;
              $$settled = false;
            }
          },
          {
            default: () => {
              return `<div class="col-header-content" data-svelte-h="svelte-1keebr9"><i${add_attribute("class", CommonHelper.getFieldTypeIcon("number"), 0)}></i> <span class="txt">Number</span></div>`;
            }
          }
        )} ${validate_component(SortHeader, "SortHeader").$$render(
          $$result,
          {
            class: "col-type-text col-field-url",
            name: "title",
            sort
          },
          {
            sort: ($$value) => {
              sort = $$value;
              $$settled = false;
            }
          },
          {
            default: () => {
              return `<div class="col-header-content" data-svelte-h="svelte-12aij12"><i${add_attribute("class", CommonHelper.getFieldTypeIcon("text"), 0)}></i> <span class="txt">Title</span></div>`;
            }
          }
        )} ${validate_component(SortHeader, "SortHeader").$$render(
          $$result,
          {
            disable: true,
            class: "col-field-author",
            name: "author",
            sort
          },
          {
            sort: ($$value) => {
              sort = $$value;
              $$settled = false;
            }
          },
          {
            default: () => {
              return `<div class="col-header-content" data-svelte-h="svelte-9ibs0t"><i class="ri-global-line"></i> <span class="txt">Author</span></div>`;
            }
          }
        )} ${validate_component(SortHeader, "SortHeader").$$render(
          $$result,
          {
            disable: true,
            class: "col-type-date col-field-created",
            name: "created",
            sort
          },
          {
            sort: ($$value) => {
              sort = $$value;
              $$settled = false;
            }
          },
          {
            default: () => {
              return `<div class="col-header-content" data-svelte-h="svelte-1kqei0j"><i${add_attribute("class", CommonHelper.getFieldTypeIcon("date"), 0)}></i> <span class="txt">Created</span></div>`;
            }
          }
        )} ${validate_component(SortHeader, "SortHeader").$$render(
          $$result,
          {
            disable: true,
            class: "col-type-date col-field-updated",
            name: "created",
            sort
          },
          {
            sort: ($$value) => {
              sort = $$value;
              $$settled = false;
            }
          },
          {
            default: () => {
              return `<div class="col-header-content" data-svelte-h="svelte-o07m0"><i${add_attribute("class", CommonHelper.getFieldTypeIcon("date"), 0)}></i> <span class="txt">Updated</span></div>`;
            }
          }
        )} <th class="col-type-action min-width"></th></tr></thead> <tbody>${items.length ? each(items, (item) => {
          return `<tr tabindex="0" class="row-handle"><td class="col-type-text col-field-method min-width"><span class="${[
            "label",
            (item.state === "closed" ? "label-danger" : "") + " " + (item.state === "open" ? "label-success" : "")
          ].join(" ").trim()}">${escape(item.state)} </span></td> <td class="col-type-number col-field-number min-width"><span class="${["label", item.status >= 400 ? "label-danger" : ""].join(" ").trim()}">${escape(item.number)} </span></td> <td class="col-type-text col-field-url"><span class="txt txt-ellipsis"${add_attribute("title", item.title, 0)}>${escape(item.title)} </span></td> <td class="col-type-text col-field-method"><span class="label">${escape(item.user)} </span></td> <td class="col-type-date col-field-created">${validate_component(FormattedDate, "FormattedDate").$$render($$result, { date: item.created_at }, {}, {})}</td> <td class="col-type-date col-field-created">${validate_component(FormattedDate, "FormattedDate").$$render($$result, { date: item.updated_at }, {}, {})}</td> <td class="col-type-action min-width" data-svelte-h="svelte-16523e2"><i class="ri-arrow-right-line"></i></td> </tr>`;
        }) : `${`<tr><td colspan="99" class="txt-center txt-hint p-xs"><h6 data-svelte-h="svelte-1rhebtm">No logs found.</h6> ${filter?.length ? `<button type="button" class="btn btn-hint btn-expanded m-t-sm" data-svelte-h="svelte-1nmdf7p"><span class="txt">Clear filters</span> </button>` : ``}</td> </tr>`}`}</tbody></table>`;
      }
    })}`;
  } while (!$$settled);
  return $$rendered;
});
const pageTitle = "Pull requests";
const Page = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let { data } = $$props;
  let { pulls } = data;
  if ($$props.data === void 0 && $$bindings.data && data !== void 0)
    $$bindings.data(data);
  return `<div class="page-wrapper"><main class="page-content"><div class="page-header-wrapper m-b-0"><header class="page-header"><nav class="breadcrumbs"><div class="breadcrumb-item">${escape(pageTitle)}</div></nav> ${validate_component(RefreshButton, "RefreshButton").$$render($$result, {}, {}, {})} <small style="opacity: 0.6" data-svelte-h="svelte-1bdfph5">Updated every hour</small> <div class="flex-fill"></div></header> ${validate_component(Searchbar, "Searchbar").$$render(
    $$result,
    {
      placeholder: "Search for PR title or number"
    },
    {},
    {}
  )} <div class="clearfix m-b-base"></div></div> ${validate_component(PRList, "PRList").$$render($$result, { items: pulls }, {}, {})}</main></div>`;
});
export {
  Page as default
};

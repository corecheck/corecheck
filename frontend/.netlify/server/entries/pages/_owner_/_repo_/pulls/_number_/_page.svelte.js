import { s as subscribe, n as null_to_empty } from "../../../../../../chunks/utils.js";
import { c as create_ssr_component, b as createEventDispatcher, d as add_attribute, a as escape, e as each, v as validate_component, m as missing_component, f as add_classes } from "../../../../../../chunks/ssr.js";
import { p as page } from "../../../../../../chunks/stores.js";
import { C as CommonHelper, S as SortHeader } from "../../../../../../chunks/SortHeader.js";
import { w as writable } from "../../../../../../chunks/index2.js";
import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime.js";
import { p as public_env } from "../../../../../../chunks/shared-server.js";
import { _ as _fetchReport, a as _fetchSonarCloudIssues } from "../../../../../../chunks/report.js";
const Toggler = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let { trigger = void 0 } = $$props;
  let { active = false } = $$props;
  let { escClose = true } = $$props;
  let { autoScroll = true } = $$props;
  let { closableClass = "closable" } = $$props;
  let { class: classes = "" } = $$props;
  let container;
  let containerChild;
  let scrollTimeoutId;
  const dispatch = createEventDispatcher();
  function hide() {
    active = false;
    clearTimeout(scrollTimeoutId);
  }
  function show() {
    active = true;
    clearTimeout(scrollTimeoutId);
    scrollTimeoutId = setTimeout(
      () => {
        if (!autoScroll) {
          return;
        }
      },
      180
    );
  }
  function toggle() {
    if (active) {
      hide();
    } else {
      show();
    }
  }
  if ($$props.trigger === void 0 && $$bindings.trigger && trigger !== void 0)
    $$bindings.trigger(trigger);
  if ($$props.active === void 0 && $$bindings.active && active !== void 0)
    $$bindings.active(active);
  if ($$props.escClose === void 0 && $$bindings.escClose && escClose !== void 0)
    $$bindings.escClose(escClose);
  if ($$props.autoScroll === void 0 && $$bindings.autoScroll && autoScroll !== void 0)
    $$bindings.autoScroll(autoScroll);
  if ($$props.closableClass === void 0 && $$bindings.closableClass && closableClass !== void 0)
    $$bindings.closableClass(closableClass);
  if ($$props.class === void 0 && $$bindings.class && classes !== void 0)
    $$bindings.class(classes);
  if ($$props.hide === void 0 && $$bindings.hide && hide !== void 0)
    $$bindings.hide(hide);
  if ($$props.show === void 0 && $$bindings.show && show !== void 0)
    $$bindings.show(show);
  if ($$props.toggle === void 0 && $$bindings.toggle && toggle !== void 0)
    $$bindings.toggle(toggle);
  {
    if (active) {
      dispatch("show");
    } else {
      dispatch("hide");
    }
  }
  return ` <div class="toggler-container" tabindex="-1"${add_attribute("this", container, 0)}>${active ? `<div class="${[escape(classes, true), active ? "active" : ""].join(" ").trim()}"${add_attribute("this", containerChild, 0)}>${slots.default ? slots.default({}) : ``}</div>` : ``}</div>`;
});
const errors = writable({});
function removeError(name) {
  errors.update((e) => {
    CommonHelper.deleteByPath(e, name);
    return e;
  });
}
const defaultError = "Invalid value";
function getErrorMessage(err) {
  if (typeof err === "object") {
    return err?.message || err?.code || defaultError;
  }
  return err || defaultError;
}
const Field = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let $errors, $$unsubscribe_errors;
  $$unsubscribe_errors = subscribe(errors, (value) => $errors = value);
  const uniqueId = "field_" + CommonHelper.randomString(7);
  let { name = "" } = $$props;
  let { inlineError = false } = $$props;
  let { class: classes = void 0 } = $$props;
  let container;
  let fieldErrors = [];
  function changed() {
    removeError(name);
  }
  if ($$props.name === void 0 && $$bindings.name && name !== void 0)
    $$bindings.name(name);
  if ($$props.inlineError === void 0 && $$bindings.inlineError && inlineError !== void 0)
    $$bindings.inlineError(inlineError);
  if ($$props.class === void 0 && $$bindings.class && classes !== void 0)
    $$bindings.class(classes);
  if ($$props.changed === void 0 && $$bindings.changed && changed !== void 0)
    $$bindings.changed(changed);
  {
    {
      fieldErrors = CommonHelper.toArray(CommonHelper.getNestedVal($errors, name));
    }
  }
  $$unsubscribe_errors();
  return ` <div class="${[escape(classes, true), fieldErrors.length ? "error" : ""].join(" ").trim()}"${add_attribute("this", container, 0)}>${slots.default ? slots.default({ uniqueId }) : ``} ${inlineError && fieldErrors.length ? `<div class="form-field-addon"><i class="ri-error-warning-fill txt-danger"></i></div>` : `${each(fieldErrors, (error) => {
    return `<div class="help-block help-block-error"><pre>${escape(getErrorMessage(error))}</pre> </div>`;
  })}`}</div>`;
});
const github = `<style>pre code.hljs{display:block;overflow-x:auto;padding:1em}code.hljs{padding:3px 5px}/*!
  Theme: GitHub
  Description: Light theme as seen on github.com
  Author: github.com
  Maintainer: @Hirse
  Updated: 2021-05-15

  Outdated base version: https://github.com/primer/github-syntax-light
  Current colors taken from GitHub's CSS
*/.hljs{color:#24292e;background:#fff}.hljs-doctag,.hljs-keyword,.hljs-meta .hljs-keyword,.hljs-template-tag,.hljs-template-variable,.hljs-type,.hljs-variable.language_{color:#d73a49}.hljs-title,.hljs-title.class_,.hljs-title.class_.inherited__,.hljs-title.function_{color:#6f42c1}.hljs-attr,.hljs-attribute,.hljs-literal,.hljs-meta,.hljs-number,.hljs-operator,.hljs-selector-attr,.hljs-selector-class,.hljs-selector-id,.hljs-variable{color:#005cc5}.hljs-meta .hljs-string,.hljs-regexp,.hljs-string{color:#032f62}.hljs-built_in,.hljs-symbol{color:#e36209}.hljs-code,.hljs-comment,.hljs-formula{color:#6a737d}.hljs-name,.hljs-quote,.hljs-selector-pseudo,.hljs-selector-tag{color:#22863a}.hljs-subst{color:#24292e}.hljs-section{color:#005cc5;font-weight:700}.hljs-bullet{color:#735c0f}.hljs-emphasis{color:#24292e;font-style:italic}.hljs-strong{color:#24292e;font-weight:700}.hljs-addition{color:#22863a;background-color:#f0fff4}.hljs-deletion{color:#b31d28;background-color:#ffeef0}</style>`;
const github$1 = github;
function defaultSearchFunc(item, search) {
  let normalizedSearch = ("" + search).replace(/\s+/g, "").toLowerCase();
  let normalizedItem = item;
  try {
    if (typeof item === "object" && item !== null) {
      normalizedItem = JSON.stringify(item);
    }
  } catch (e) {
  }
  return ("" + normalizedItem).replace(/\s+/g, "").toLowerCase().includes(normalizedSearch);
}
const Select = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let filteredItems;
  let isSelected;
  let { id = "" } = $$props;
  let { noOptionsText = "No options found" } = $$props;
  let { selectPlaceholder = "- Select -" } = $$props;
  let { searchPlaceholder = "Search..." } = $$props;
  let { items = [] } = $$props;
  let { multiple = false } = $$props;
  let { disabled = false } = $$props;
  let { readonly = false } = $$props;
  let { selected = multiple ? [] : void 0 } = $$props;
  let { toggle = multiple } = $$props;
  let { closable = true } = $$props;
  let { labelComponent = void 0 } = $$props;
  let { labelComponentProps = {} } = $$props;
  let { optionComponent = void 0 } = $$props;
  let { optionComponentProps = {} } = $$props;
  let { searchable = false } = $$props;
  let { searchFunc = void 0 } = $$props;
  let { class: classes = "" } = $$props;
  let toggler;
  let searchTerm = "";
  let container = void 0;
  let labelDiv = void 0;
  function deselectItem(item) {
    if (CommonHelper.isEmpty(selected)) {
      return;
    }
    let normalized = CommonHelper.toArray(selected);
    if (CommonHelper.inArray(normalized, item)) {
      CommonHelper.removeByValue(normalized, item);
      selected = normalized;
    }
  }
  function selectItem(item) {
    if (multiple) {
      let normalized = CommonHelper.toArray(selected);
      if (!CommonHelper.inArray(normalized, item)) {
        selected = [...normalized, item];
      }
    } else {
      selected = item;
    }
  }
  function toggleItem(item) {
    return isSelected(item) ? deselectItem(item) : selectItem(item);
  }
  function reset() {
    selected = multiple ? [] : void 0;
  }
  function showDropdown() {
    toggler?.show && toggler?.show();
  }
  function hideDropdown() {
    toggler?.hide && toggler?.hide();
  }
  function ensureSelectedExist() {
    if (CommonHelper.isEmpty(selected) || CommonHelper.isEmpty(items)) {
      return;
    }
    let selectedArray = CommonHelper.toArray(selected);
    let unselectedArray = [];
    for (const selectedItem of selectedArray) {
      if (!CommonHelper.inArray(items, selectedItem)) {
        unselectedArray.push(selectedItem);
      }
    }
    if (unselectedArray.length) {
      for (const item of unselectedArray) {
        CommonHelper.removeByValue(selectedArray, item);
      }
      selected = multiple ? selectedArray : selectedArray[0];
    }
  }
  function resetSearch() {
    searchTerm = "";
  }
  function filterItems(items2, search) {
    items2 = items2 || [];
    const filterFunc = searchFunc || defaultSearchFunc;
    return items2.filter((item) => filterFunc(item, search)) || [];
  }
  if ($$props.id === void 0 && $$bindings.id && id !== void 0)
    $$bindings.id(id);
  if ($$props.noOptionsText === void 0 && $$bindings.noOptionsText && noOptionsText !== void 0)
    $$bindings.noOptionsText(noOptionsText);
  if ($$props.selectPlaceholder === void 0 && $$bindings.selectPlaceholder && selectPlaceholder !== void 0)
    $$bindings.selectPlaceholder(selectPlaceholder);
  if ($$props.searchPlaceholder === void 0 && $$bindings.searchPlaceholder && searchPlaceholder !== void 0)
    $$bindings.searchPlaceholder(searchPlaceholder);
  if ($$props.items === void 0 && $$bindings.items && items !== void 0)
    $$bindings.items(items);
  if ($$props.multiple === void 0 && $$bindings.multiple && multiple !== void 0)
    $$bindings.multiple(multiple);
  if ($$props.disabled === void 0 && $$bindings.disabled && disabled !== void 0)
    $$bindings.disabled(disabled);
  if ($$props.readonly === void 0 && $$bindings.readonly && readonly !== void 0)
    $$bindings.readonly(readonly);
  if ($$props.selected === void 0 && $$bindings.selected && selected !== void 0)
    $$bindings.selected(selected);
  if ($$props.toggle === void 0 && $$bindings.toggle && toggle !== void 0)
    $$bindings.toggle(toggle);
  if ($$props.closable === void 0 && $$bindings.closable && closable !== void 0)
    $$bindings.closable(closable);
  if ($$props.labelComponent === void 0 && $$bindings.labelComponent && labelComponent !== void 0)
    $$bindings.labelComponent(labelComponent);
  if ($$props.labelComponentProps === void 0 && $$bindings.labelComponentProps && labelComponentProps !== void 0)
    $$bindings.labelComponentProps(labelComponentProps);
  if ($$props.optionComponent === void 0 && $$bindings.optionComponent && optionComponent !== void 0)
    $$bindings.optionComponent(optionComponent);
  if ($$props.optionComponentProps === void 0 && $$bindings.optionComponentProps && optionComponentProps !== void 0)
    $$bindings.optionComponentProps(optionComponentProps);
  if ($$props.searchable === void 0 && $$bindings.searchable && searchable !== void 0)
    $$bindings.searchable(searchable);
  if ($$props.searchFunc === void 0 && $$bindings.searchFunc && searchFunc !== void 0)
    $$bindings.searchFunc(searchFunc);
  if ($$props.class === void 0 && $$bindings.class && classes !== void 0)
    $$bindings.class(classes);
  if ($$props.deselectItem === void 0 && $$bindings.deselectItem && deselectItem !== void 0)
    $$bindings.deselectItem(deselectItem);
  if ($$props.selectItem === void 0 && $$bindings.selectItem && selectItem !== void 0)
    $$bindings.selectItem(selectItem);
  if ($$props.toggleItem === void 0 && $$bindings.toggleItem && toggleItem !== void 0)
    $$bindings.toggleItem(toggleItem);
  if ($$props.reset === void 0 && $$bindings.reset && reset !== void 0)
    $$bindings.reset(reset);
  if ($$props.showDropdown === void 0 && $$bindings.showDropdown && showDropdown !== void 0)
    $$bindings.showDropdown(showDropdown);
  if ($$props.hideDropdown === void 0 && $$bindings.hideDropdown && hideDropdown !== void 0)
    $$bindings.hideDropdown(hideDropdown);
  let $$settled;
  let $$rendered;
  do {
    $$settled = true;
    {
      if (items) {
        ensureSelectedExist();
        resetSearch();
      }
    }
    filteredItems = filterItems(items, searchTerm);
    isSelected = function(item) {
      const normalized = CommonHelper.toArray(selected);
      return CommonHelper.inArray(normalized, item);
    };
    $$rendered = `<div class="${[
      "select " + escape(classes, true),
      (multiple ? "multiple" : "") + " " + (disabled ? "disabled" : "") + " " + (readonly ? "readonly" : "")
    ].join(" ").trim()}"${add_attribute("this", container, 0)}> <div${add_attribute("tabindex", disabled || readonly ? "-1" : "0", 0)} class="${[
      "selected-container",
      (disabled ? "disabled" : "") + " " + (readonly ? "readonly" : "")
    ].join(" ").trim()}"${add_attribute("this", labelDiv, 0)}>${CommonHelper.toArray(selected).length ? each(CommonHelper.toArray(selected), (item, i) => {
      return `<div class="option">${labelComponent ? `${validate_component(labelComponent || missing_component, "svelte:component").$$render($$result, Object.assign({}, { item }, labelComponentProps), {}, {})}` : `<span class="txt">${escape(item)}</span>`} ${multiple || toggle ? ` <span class="clear" data-svelte-h="svelte-11pdsul"><i class="ri-close-line"></i> </span>` : ``} </div>`;
    }) : `<div class="${["block txt-placeholder", !disabled && !readonly ? "link-hint" : ""].join(" ").trim()}">${escape(selectPlaceholder)} </div>`}</div> ${!disabled && !readonly ? `${validate_component(Toggler, "Toggler").$$render(
      $$result,
      {
        class: "dropdown dropdown-block options-dropdown dropdown-left",
        trigger: labelDiv,
        this: toggler
      },
      {
        this: ($$value) => {
          toggler = $$value;
          $$settled = false;
        }
      },
      {
        default: () => {
          return `${searchable ? `<div class="form-field form-field-sm options-search"><label class="input-group"><div class="addon p-r-0" data-svelte-h="svelte-1xvh8fb"><i class="ri-search-line"></i></div>  <input autofocus type="text"${add_attribute("placeholder", searchPlaceholder, 0)}${add_attribute("value", searchTerm, 0)}> ${searchTerm.length ? `<div class="addon suffix p-r-5"><button type="button" class="btn btn-sm btn-circle btn-transparent clear" data-svelte-h="svelte-1u1nquo"><i class="ri-close-line"></i></button></div>` : ``}</label></div>` : ``} ${slots.beforeOptions ? slots.beforeOptions({}) : ``} <div class="options-list">${filteredItems.length ? each(filteredItems, (item) => {
            return ` <div tabindex="0" class="${[
              "dropdown-item option",
              (closable ? "closable" : "") + " " + (isSelected(item) ? "selected" : "")
            ].join(" ").trim()}">${optionComponent ? `${validate_component(optionComponent || missing_component, "svelte:component").$$render($$result, Object.assign({}, { item }, optionComponentProps), {}, {})}` : `${escape(item)}`} </div>`;
          }) : `${noOptionsText ? `<div class="txt-missing">${escape(noOptionsText)}</div>` : ``}`}</div> ${slots.afterOptions ? slots.afterOptions({}) : ``}`;
        }
      }
    )}` : ``}</div>`;
  } while (!$$settled);
  return $$rendered;
});
const CoverageReportSelectOption = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  dayjs.extend(relativeTime);
  let { item } = $$props;
  if ($$props.item === void 0 && $$bindings.item && item !== void 0)
    $$bindings.item(item);
  return `<span class="dropdown-item-text">${escape(item.commit.substring(0, 7))} - ${escape(dayjs().to(dayjs(item.created_at)))} ${escape(item.status === "pending" ? "(PENDING)" : "")}</span>`;
});
const Accordion = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  const dispatch = createEventDispatcher();
  let accordionElem;
  let expandTimeoutId;
  let { class: classes = "" } = $$props;
  let { draggable = false } = $$props;
  let { active = false } = $$props;
  let { interactive = true } = $$props;
  let { single = false } = $$props;
  function isExpanded() {
    return !!active;
  }
  function expand() {
    collapseSiblings();
    active = true;
    dispatch("expand");
  }
  function collapse() {
    active = false;
    clearTimeout(expandTimeoutId);
    dispatch("collapse");
  }
  function toggle() {
    dispatch("toggle");
    if (active) {
      collapse();
    } else {
      expand();
    }
  }
  function collapseSiblings() {
    if (single && accordionElem.closest(".accordions")) {
      const handlers = accordionElem.closest(".accordions").querySelectorAll(".accordion.active .accordion-header.interactive");
      for (const handler of handlers) {
        handler.click();
      }
    }
  }
  if ($$props.class === void 0 && $$bindings.class && classes !== void 0)
    $$bindings.class(classes);
  if ($$props.draggable === void 0 && $$bindings.draggable && draggable !== void 0)
    $$bindings.draggable(draggable);
  if ($$props.active === void 0 && $$bindings.active && active !== void 0)
    $$bindings.active(active);
  if ($$props.interactive === void 0 && $$bindings.interactive && interactive !== void 0)
    $$bindings.interactive(interactive);
  if ($$props.single === void 0 && $$bindings.single && single !== void 0)
    $$bindings.single(single);
  if ($$props.isExpanded === void 0 && $$bindings.isExpanded && isExpanded !== void 0)
    $$bindings.isExpanded(isExpanded);
  if ($$props.expand === void 0 && $$bindings.expand && expand !== void 0)
    $$bindings.expand(expand);
  if ($$props.collapse === void 0 && $$bindings.collapse && collapse !== void 0)
    $$bindings.collapse(collapse);
  if ($$props.toggle === void 0 && $$bindings.toggle && toggle !== void 0)
    $$bindings.toggle(toggle);
  if ($$props.collapseSiblings === void 0 && $$bindings.collapseSiblings && collapseSiblings !== void 0)
    $$bindings.collapseSiblings(collapseSiblings);
  return `<div class="${[
    "accordion " + escape("", true) + " " + escape(classes, true),
    active ? "active" : ""
  ].join(" ").trim()}"${add_attribute("this", accordionElem, 0)}><button type="button" class="${["accordion-header", interactive ? "interactive" : ""].join(" ").trim()}"${add_attribute("draggable", draggable, 0)}>${slots.header ? slots.header({ active }) : ``}</button> ${active ? `<div class="accordion-content">${slots.default ? slots.default({}) : ``}</div>` : ``}</div>`;
});
const Coverage_svelte_svelte_type_style_lang = "";
const css$2 = {
  code: ".full-width.svelte-1ksyllb.svelte-1ksyllb{width:100% !important}.code.svelte-1ksyllb.svelte-1ksyllb{background-color:#f1f1f1;padding:1rem;border-radius:0.25rem;margin-bottom:1rem;overflow:auto}.filename.svelte-1ksyllb.svelte-1ksyllb{margin-bottom:0.5rem;background-color:#2f2f30;color:#fff;padding:0.1rem 0.5rem;font-size:0.8rem;border-radius:0.5rem;display:inline-block}.line.svelte-1ksyllb.svelte-1ksyllb{display:flex;align-items:center}.line.svelte-1ksyllb .line-changed-covered.svelte-1ksyllb{background-color:#66ee86}.line.svelte-1ksyllb .line-changed-uncovered.svelte-1ksyllb{background-color:#e97373}.line.svelte-1ksyllb .line-unchanged-covered.svelte-1ksyllb{background-color:#bbf1c8}.line.svelte-1ksyllb .line-unchanged-uncovered.svelte-1ksyllb{background-color:#f1c7c7}.line-number.svelte-1ksyllb.svelte-1ksyllb{text-align:right;margin-right:0.5rem;border-right:1px solid #d0d5db}.context-button.svelte-1ksyllb.svelte-1ksyllb{position:absolute;top:0px;right:15px}pre code.hljs{white-space:pre-wrap}",
  map: null
};
const Coverage = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let { coverage } = $$props;
  let { name } = $$props;
  let { description } = $$props;
  let { icon } = $$props;
  let { color } = $$props;
  const randomKey = Math.random().toString(36).replace(/[^a-z]+/g, "").substring(0, 5);
  if ($$props.coverage === void 0 && $$bindings.coverage && coverage !== void 0)
    $$bindings.coverage(coverage);
  if ($$props.name === void 0 && $$bindings.name && name !== void 0)
    $$bindings.name(name);
  if ($$props.description === void 0 && $$bindings.description && description !== void 0)
    $$bindings.description(description);
  if ($$props.icon === void 0 && $$bindings.icon && icon !== void 0)
    $$bindings.icon(icon);
  if ($$props.color === void 0 && $$bindings.color && color !== void 0)
    $$bindings.color(color);
  $$result.css.add(css$2);
  return `<div class="flex"><div><div class="flex"><h1>${escape(name)} <i class="${escape(null_to_empty(`${color} ${icon}`), true) + " svelte-1ksyllb"}"></i></h1> ${coverage ? `${`<button type="button" class="btn btn-sm btn-primary" data-svelte-h="svelte-g52fu1"><i class="ri-arrow-down-s-line"></i>
                        Expand all</button>`}` : ``}</div> <p>${escape(description)}</p></div></div> <div class="clearfix m-b-base"></div> ${!coverage ? `<div class="alert alert-info" style="text-align: center" data-svelte-h="svelte-1lsvvcf"><i class="ri-information-line"></i> No coverage data</div>` : `<div class="${escape(null_to_empty(`accordions ${randomKey}`), true) + " svelte-1ksyllb"}">${each(Object.keys(coverage), (filename) => {
    return `${validate_component(Accordion, "Accordion").$$render($$result, {}, {}, {
      header: () => {
        return `<div class="inline-flex"><i class="ri-file-code-line"></i> <span class="txt">${escape(filename)}</span></div> <div class="flex-fill"></div> `;
      },
      default: () => {
        return `${each(coverage[filename], (hunk) => {
          return `<pre><div class="code svelte-1ksyllb">${each(hunk.lines, (line) => {
            return `<div class="line svelte-1ksyllb"><a target="_blank" class="line-number link-primary txt-mono svelte-1ksyllb">${escape(line.line_number)} </a><span class="${[
              "txt-mono svelte-1ksyllb",
              (line.tested && line.highlight && line.covered ? "line-changed-covered" : "") + " " + (line.tested && line.covered && line.context ? "line-unchanged-covered" : "") + " " + (line.tested && line.highlight && !line.covered ? "line-changed-uncovered" : "") + " " + (line.tested && !line.covered && line.context ? "line-unchanged-uncovered" : "")
            ].join(" ").trim()}">${escape(line.content)}</span></div>`;
          })}</div></pre>`;
        })} `;
      }
    })}`;
  })}</div>`} <div class="clearfix m-b-base"></div>`;
});
const Sonarcloud_svelte_svelte_type_style_lang = "";
const css$1 = {
  code: ".highlight.svelte-14qf9o0{background-color:#e97373}.code.svelte-14qf9o0{background-color:#f1f1f1;padding:1rem;border-radius:0.25rem;margin-bottom:1rem;overflow:auto}.line.svelte-14qf9o0{display:flex;align-items:center}.line-number.svelte-14qf9o0{text-align:right;margin-right:0.5rem;border-right:1px solid #d0d5db}pre code.hljs{white-space:pre-wrap}",
  map: null
};
const Sonarcloud = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let { report } = $$props;
  let { issues = [] } = $$props;
  if ($$props.report === void 0 && $$bindings.report && report !== void 0)
    $$bindings.report(report);
  if ($$props.issues === void 0 && $$bindings.issues && issues !== void 0)
    $$bindings.issues(issues);
  $$result.css.add(css$1);
  return `<h1 data-svelte-h="svelte-1mbvfmc">Sonarcloud</h1> <div class="clearfix m-b-base"></div> <div class="accordions sonarcloud">${each(
    issues.sort((a, b) => {
      if (a.severity === "BLOCKER")
        return -1;
      if (b.severity === "BLOCKER")
        return 1;
      if (a.severity === "CRITICAL")
        return -1;
      if (b.severity === "CRITICAL")
        return 1;
      if (a.severity === "MAJOR")
        return -1;
      if (b.severity === "MAJOR")
        return 1;
      if (a.severity === "MINOR")
        return -1;
      if (b.severity === "MINOR")
        return 1;
      if (a.severity === "INFO")
        return -1;
      if (b.severity === "INFO")
        return 1;
      return 0;
    }),
    (issue) => {
      return `${validate_component(Accordion, "Accordion").$$render($$result, {}, {}, {
        header: () => {
          return `<div class="inline-flex"><i class="ri-file-code-line"></i> <span class="txt">${escape(issue.message)} </span></div> <div class="flex-fill"></div> <span class="${[
            "label",
            (issue.severity === "BLOCKER" ? "label-success" : "") + " " + (issue.severity === "CRITICAL" ? "label-warning" : "") + " " + (issue.severity === "MAJOR" ? "label-danger" : "") + " " + (issue.severity === "MINOR" ? "label-info" : "") + " " + (issue.severity === "INFO" ? "label-secondary" : "")
          ].join(" ").trim()}">${escape(issue.severity)}</span> `;
        },
        default: () => {
          return `<div class="form-field m-b-0"><pre><div class="code svelte-14qf9o0"><span class="filename">${escape(issue.component.split(":")[1])}</span>
${each(issue.sources, (line) => {
            return `<div class="line svelte-14qf9o0"><span class="line-number txt-mono svelte-14qf9o0">${escape(line.line)} </span><span class="${[
              "txt-mono svelte-14qf9o0",
              line.line >= issue.textRange.startLine && line.line <= issue.textRange.endLine ? "highlight" : ""
            ].join(" ").trim()}"><!-- HTML_TAG_START -->${line.code}<!-- HTML_TAG_END --></span></div>`;
          })}</div></pre>  <a class="btn btn-primary btn-sm" href="${"https://sonarcloud.io/project/issues?id=aureleoules_bitcoin&branch=" + escape(report.pr_number, true) + "-" + escape(report.commit, true) + "&resolved=false&open=" + escape(issue.key, true)}" target="_blank">Open in SonarCloud</a></div> `;
        }
      })}`;
    }
  )} </div>`;
});
const threshold = 0.1;
function displayBenchNumber(n, showSign = false) {
  if (!n)
    return 0;
  return Math.round(n).toLocaleString("en-US", {
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  });
}
function displayPercentage(n) {
  const r = (n * 100).toLocaleString("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  });
  if (n > 0)
    return "+" + r;
  return r;
}
const Benchmarks = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let sort = "-diff";
  let { report } = $$props;
  let showOnlySignificant = true;
  function getUnit(benchmark) {
    if (!report.benchmarks_grouped[benchmark])
      return "";
    return report.benchmarks_grouped[benchmark]["unit"];
  }
  function getNsPerUnit(benchmark) {
    if (!report.benchmarks_grouped[benchmark])
      return 0;
    return report.benchmarks_grouped[benchmark]["median(elapsed)"] / (1e-9 * report.benchmarks_grouped[benchmark]["batch"]);
  }
  function getNsPerUnitMaster(benchmark) {
    if (!report.base_report.benchmarks_grouped[benchmark])
      return 0;
    return report.base_report.benchmarks_grouped[benchmark]["median(elapsed)"] / (1e-9 * report.base_report.benchmarks_grouped[benchmark]["batch"]);
  }
  function getNsPerUnitDiff(benchmark) {
    if (!report.base_report.benchmarks_grouped[benchmark])
      return 0;
    return (getNsPerUnit(benchmark) - getNsPerUnitMaster(benchmark)) / getNsPerUnitMaster(benchmark);
  }
  function getUnitPerSecond(benchmark) {
    if (!report.benchmarks_grouped[benchmark])
      return 0;
    return report.benchmarks_grouped[benchmark]["batch"] / report.benchmarks_grouped[benchmark]["median(elapsed)"];
  }
  function getUnitPerSecondMaster(benchmark) {
    if (!report.base_report.benchmarks_grouped[benchmark])
      return 0;
    return report.base_report.benchmarks_grouped[benchmark]["batch"] / report.base_report.benchmarks_grouped[benchmark]["median(elapsed)"];
  }
  function getUnitPerSecondDiff(benchmark) {
    if (!report.base_report.benchmarks_grouped[benchmark])
      return 0;
    return (getUnitPerSecond(benchmark) - getUnitPerSecondMaster(benchmark)) / getUnitPerSecondMaster(benchmark);
  }
  function getIPC(benchmark) {
    if (!report.benchmarks_grouped[benchmark])
      return 0;
    return report.benchmarks_grouped[benchmark]["median(instructions)"] / report.benchmarks_grouped[benchmark]["median(cpucycles)"];
  }
  function getIPCMaster(benchmark) {
    if (!report.base_report.benchmarks_grouped[benchmark])
      return 0;
    return report.base_report.benchmarks_grouped[benchmark]["median(instructions)"] / report.base_report.benchmarks_grouped[benchmark]["median(cpucycles)"];
  }
  function getIPCDiff(benchmark) {
    if (!report.base_report.benchmarks_grouped[benchmark])
      return 0;
    return (getIPC(benchmark) - getIPCMaster(benchmark)) / getIPCMaster(benchmark);
  }
  function getCyclesPerUnit(benchmark) {
    if (!report.benchmarks_grouped[benchmark])
      return 0;
    return report.benchmarks_grouped[benchmark]["median(cpucycles)"] / report.benchmarks_grouped[benchmark]["batch"];
  }
  function getCyclesPerUnitMaster(benchmark) {
    if (!report.base_report.benchmarks_grouped[benchmark])
      return 0;
    return report.base_report.benchmarks_grouped[benchmark]["median(cpucycles)"] / report.base_report.benchmarks_grouped[benchmark]["batch"];
  }
  function getCyclesPerUnitDiff(benchmark) {
    if (!report.base_report.benchmarks_grouped[benchmark])
      return 0;
    return (getCyclesPerUnit(benchmark) - getCyclesPerUnitMaster(benchmark)) / getCyclesPerUnitMaster(benchmark);
  }
  function getInstructionsPerUnit(benchmark) {
    if (!report.benchmarks_grouped[benchmark])
      return 0;
    return report.benchmarks_grouped[benchmark]["median(instructions)"] / report.benchmarks_grouped[benchmark]["batch"];
  }
  function getInstructionsPerUnitMaster(benchmark) {
    if (!report.base_report.benchmarks_grouped[benchmark])
      return 0;
    return report.base_report.benchmarks_grouped[benchmark]["median(instructions)"] / report.base_report.benchmarks_grouped[benchmark]["batch"];
  }
  function getInstructionsPerUnitDiff(benchmark) {
    if (!report.base_report.benchmarks_grouped[benchmark])
      return 0;
    return (getInstructionsPerUnit(benchmark) - getInstructionsPerUnitMaster(benchmark)) / getInstructionsPerUnitMaster(benchmark);
  }
  function getBranchesPerUnit(benchmark) {
    if (!report.benchmarks_grouped[benchmark])
      return 0;
    return report.benchmarks_grouped[benchmark]["median(branchinstructions)"] / report.benchmarks_grouped[benchmark]["batch"];
  }
  function getBranchesPerUnitMaster(benchmark) {
    if (!report.base_report.benchmarks_grouped[benchmark])
      return 0;
    return report.base_report.benchmarks_grouped[benchmark]["median(branchinstructions)"] / report.base_report.benchmarks_grouped[benchmark]["batch"];
  }
  function getBranchesPerUnitDiff(benchmark) {
    return (getBranchesPerUnit(benchmark) - getBranchesPerUnitMaster(benchmark)) / getBranchesPerUnitMaster(benchmark);
  }
  function isSignificant(benchmark) {
    return Math.abs(getNsPerUnitDiff(benchmark)) > threshold || Math.abs(getUnitPerSecondDiff(benchmark)) > threshold || Math.abs(getInstructionsPerUnitDiff(benchmark)) > threshold || Math.abs(getCyclesPerUnitDiff(benchmark)) > threshold || Math.abs(getIPCDiff(benchmark)) > threshold || Math.abs(getBranchesPerUnitDiff(benchmark)) > threshold;
  }
  if ($$props.report === void 0 && $$bindings.report && report !== void 0)
    $$bindings.report(report);
  let $$settled;
  let $$rendered;
  do {
    $$settled = true;
    $$rendered = `<div class="flex" data-svelte-h="svelte-j9940p"><h1>Benchmarks <span class="label label-success">Beta</span></h1></div> <div class="clearfix m-b-base"></div> ${report.benchmark_status === "pending" ? `<div class="alert alert-warning" style="text-align: center" data-svelte-h="svelte-9aek69"><i class="ri-information-line"></i> Benchmarks are currently being generated,
        please come back later.</div>` : `${report.benchmark_status === "failure" ? `<div class="alert alert-danger" style="text-align: center" data-svelte-h="svelte-1vhnac2"><i class="ri-information-line"></i> An error occured while generating the benchmarks. Push a new commit to re-run the benchmarks.</div>` : `${report.benchmark_status === "not_found" ? `<div class="alert alert-info" style="text-align: center" data-svelte-h="svelte-oalnbb"><i class="ri-information-line"></i> No benchmarks available. Push a new commit to re-run the benchmarks.</div>` : `${report.benchmark_status === "success" ? `${validate_component(Field, "Field").$$render(
      $$result,
      {
        class: "form-field form-field-toggle",
        name: "verified"
      },
      {},
      {
        default: ({ uniqueId }) => {
          return `<input type="checkbox"${add_attribute("id", uniqueId, 0)}${add_attribute("checked", showOnlySignificant, 1)}> <label${add_attribute("for", uniqueId, 0)}>Only show benchmarks with a significant difference</label>`;
        }
      }
    )} <table class="table"><thead><tr>${validate_component(SortHeader, "SortHeader").$$render(
      $$result,
      { name: "name", sort },
      {
        sort: ($$value) => {
          sort = $$value;
          $$settled = false;
        }
      },
      {
        default: () => {
          return `<div class="col-header-content" data-svelte-h="svelte-p8g8f0"><i class="ri-text"></i> <span class="txt">Name</span></div>`;
        }
      }
    )} ${validate_component(SortHeader, "SortHeader").$$render(
      $$result,
      {
        class: "col-type-number col-field-type",
        name: "ns/unit",
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
          return `<div class="col-header-content" data-svelte-h="svelte-171x0j2"><i class="ri-percent-line"></i> <span class="txt">ns/unit</span></div>`;
        }
      }
    )} ${validate_component(SortHeader, "SortHeader").$$render(
      $$result,
      {
        class: "col-type-number col-field-type",
        name: "unit/s",
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
          return `<div class="col-header-content" data-svelte-h="svelte-1kmz2sg"><i class="ri-percent-line"></i> <span class="txt">unit/s</span></div>`;
        }
      }
    )} ${validate_component(SortHeader, "SortHeader").$$render(
      $$result,
      {
        class: "col-type-number col-field-type",
        name: "ins/unit",
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
          return `<div class="col-header-content" data-svelte-h="svelte-1igpwc8"><i class="ri-cpu-line"></i> <span class="txt">ins/unit</span></div>`;
        }
      }
    )} ${validate_component(SortHeader, "SortHeader").$$render(
      $$result,
      {
        class: "col-type-number col-field-type",
        name: "cyc/unit",
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
          return `<div class="col-header-content" data-svelte-h="svelte-1ojp45j"><i class="ri-database-2-line"></i> <span class="txt">cyc/unit</span></div>`;
        }
      }
    )} ${validate_component(SortHeader, "SortHeader").$$render(
      $$result,
      {
        class: "col-type-number col-field-type",
        name: "ipc",
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
          return `<div class="col-header-content" data-svelte-h="svelte-1byq79h"><i class="ri-database-2-line"></i> <span class="txt">IPC</span></div>`;
        }
      }
    )} ${validate_component(SortHeader, "SortHeader").$$render(
      $$result,
      {
        class: "col-type-number col-field-type",
        name: "totalTime",
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
          return `<div class="col-header-content" data-svelte-h="svelte-1wdz80n"><i class="ri-time-line"></i> <span class="txt">Total time</span></div>`;
        }
      }
    )}</tr></thead> <tbody>${each(
      Object.keys(report.benchmarks_grouped).sort((a, b) => {
        const benchA = report.benchmarks_grouped[a];
        const benchB = report.benchmarks_grouped[b];
        if (sort === "+name")
          return benchA.name.localeCompare(benchB.name);
        if (sort === "-name")
          return benchB.name.localeCompare(benchA.name);
        if (sort === "+ns/unit")
          return getNsPerUnit(a) - getNsPerUnit(b);
        if (sort === "-ns/unit")
          return getNsPerUnit(b) - getNsPerUnit(a);
        if (sort === "+unit/s")
          return getUnitPerSecond(a) - getUnitPerSecond(b);
        if (sort === "-unit/s")
          return getUnitPerSecond(b) - getUnitPerSecond(a);
        if (sort === "+ins/unit")
          return getInstructionsPerUnit(a) - getInstructionsPerUnit(b);
        if (sort === "-ins/unit")
          return getInstructionsPerUnit(b) - getInstructionsPerUnit(a);
        if (sort === "+cyc/unit")
          return getCyclesPerUnit(a) - getCyclesPerUnit(b);
        if (sort === "-cyc/unit")
          return getCyclesPerUnit(b) - getCyclesPerUnit(a);
        if (sort === "+ipc")
          return getIPC(a) - getIPC(b);
        if (sort === "-ipc")
          return getIPC(b) - getIPC(a);
        if (sort === "+bra/unit")
          return getBranchesPerUnit(a) - getBranchesPerUnit(b);
        if (sort === "-bra/unit")
          return getBranchesPerUnit(b) - getBranchesPerUnit(a);
        return 0;
      }).filter((b) => ![
        "AddrManSelectFromAlmostEmpty",
        "RollingBloomReset",
        "AddrManGetAddr",
        "LoadExternalBlockFile",
        "PrevectorDeserializeNontrivial",
        "GCSFilterConstruct"
      ].includes(b)).filter((b) => isSignificant(b)),
      (benchmark) => {
        return `<tr><td class="col-field-id"><p>${escape(report.benchmarks_grouped[benchmark].name.length > 50 ? report.benchmarks_grouped[benchmark].name.substring(0, 50) + "..." : report.benchmarks_grouped[benchmark].name)} </p></td> <td class="col-type-number col-field-pr"><p>${escape(displayBenchNumber(getNsPerUnit(benchmark)))} <small class="txt-hint">ns/${escape(getUnit(benchmark))}</small> <small${add_classes(((getNsPerUnitDiff(benchmark) < -threshold ? "txt-success" : "") + " " + (getNsPerUnitDiff(benchmark) > threshold ? "txt-danger" : "") + " " + (getNsPerUnitDiff(benchmark) >= -threshold && getNsPerUnitDiff(benchmark) <= threshold ? "txt-hint" : "")).trim())}>${escape(displayPercentage(getNsPerUnitDiff(benchmark)))}%</small> </p></td> <td class="col-type-number col-field-pr"><p>${escape(displayBenchNumber(getUnitPerSecond(benchmark)))} <small class="txt-hint">${escape(getUnit(benchmark))}/s</small> <small${add_classes(((getUnitPerSecondDiff(benchmark) > threshold ? "txt-success" : "") + " " + (getUnitPerSecondDiff(benchmark) < -threshold ? "txt-danger" : "") + " " + (getUnitPerSecondDiff(benchmark) <= threshold && getUnitPerSecondDiff(benchmark) >= -threshold ? "txt-hint" : "")).trim())}>${escape(displayPercentage(getUnitPerSecondDiff(benchmark)))}%</small> </p></td> <td class="col-type-number col-field-pr"><p>${escape(displayBenchNumber(getInstructionsPerUnit(benchmark)))} <small class="txt-hint">ins/${escape(getUnit(benchmark))}</small> <small${add_classes(((getInstructionsPerUnitDiff(benchmark) < -threshold ? "txt-success" : "") + " " + (getInstructionsPerUnitDiff(benchmark) > threshold ? "txt-danger" : "") + " " + (getInstructionsPerUnitDiff(benchmark) >= -threshold && getInstructionsPerUnitDiff(benchmark) <= threshold ? "txt-hint" : "")).trim())}>${escape(displayPercentage(getInstructionsPerUnitDiff(benchmark)))}%</small> </p></td> <td class="col-type-number col-field-pr"><p>${escape(displayBenchNumber(getCyclesPerUnit(benchmark)))} <small class="txt-hint">cyc/${escape(getUnit(benchmark))}</small> <small${add_classes(((getCyclesPerUnitDiff(benchmark) < -threshold ? "txt-success" : "") + " " + (getCyclesPerUnitDiff(benchmark) > threshold ? "txt-danger" : "") + " " + (getCyclesPerUnitDiff(benchmark) >= -threshold && getCyclesPerUnitDiff(benchmark) <= threshold ? "txt-hint" : "")).trim())}>${escape(displayPercentage(getCyclesPerUnitDiff(benchmark)))}%</small> </p></td> <td class="col-type-number col-field-pr"><p>${escape(getIPC(benchmark).toLocaleString("en-US", {
          minimumFractionDigits: 2,
          maximumFractionDigits: 2
        }))} <small class="txt-hint" data-svelte-h="svelte-1hh4kvt">IPC</small> <small${add_classes(((getIPCDiff(benchmark) < -threshold ? "txt-danger" : "") + " " + (getIPCDiff(benchmark) > threshold ? "txt-success" : "") + " " + (getIPCDiff(benchmark) >= -threshold && getIPCDiff(benchmark) <= threshold ? "txt-hint" : "")).trim())}>${escape(displayPercentage(getIPCDiff(benchmark)))}%</small> </p></td> <td class="col-type-number col-field-pr"><p>${escape(report.benchmarks_grouped[benchmark]["totalTime"].toLocaleString("en-US", {
          minimumFractionDigits: 2,
          maximumFractionDigits: 2
        }))} <small class="txt-hint" data-svelte-h="svelte-1x59rac">seconds</small> </p></td> </tr>`;
      }
    )}</tbody></table>` : ``}`}`}`}`;
  } while (!$$settled);
  return $$rendered;
});
const _page_svelte_svelte_type_style_lang = "";
const css = {
  code: "@media(max-width: 1100px){.cov-container.svelte-1nzmi39.svelte-1nzmi39{flex-direction:column}.cov-container.svelte-1nzmi39 .cov-col.svelte-1nzmi39{width:100% !important}}.cov-container.svelte-1nzmi39 .cov-col.svelte-1nzmi39{width:48%}",
  map: null
};
const pageTitle = "Pull requests";
const Page = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let $page, $$unsubscribe_page;
  $$unsubscribe_page = subscribe(page, (value) => $page = value);
  let { data } = $$props;
  let { pr, sonarcloud, report } = data;
  console.log(report);
  let selectedReport = pr.reports[0];
  let prev = selectedReport;
  let fetching = false;
  if ($$props.data === void 0 && $$bindings.data && data !== void 0)
    $$bindings.data(data);
  $$result.css.add(css);
  let $$settled;
  let $$rendered;
  do {
    $$settled = true;
    {
      {
        (async () => {
          if (selectedReport.id !== prev.id) {
            prev = selectedReport;
            fetching = true;
            report = await _fetchReport(fetch, public_env.PUBLIC_ENDPOINT, pr.number, selectedReport.id);
            sonarcloud = await await _fetchSonarCloudIssues(fetch, report.pr_number, report.commit);
            fetching = false;
          }
        })();
      }
    }
    $$rendered = `${$$result.head += `<!-- HEAD_svelte-1wyg5b0_START --><!-- HTML_TAG_START -->${github$1}<!-- HTML_TAG_END --><!-- HEAD_svelte-1wyg5b0_END -->`, ""} <div class="page-wrapper"><main class="page-content"><div class="page-header-wrapper m-b-0"><header class="page-header"><nav class="breadcrumbs"><div class="breadcrumb-item">${escape(pageTitle)}</div> <div class="breadcrumb-item">#${escape($page.params.number)}</div></nav> <div class="flex-fill"></div></header> <h1 class="flex"><a${add_attribute("href", "https://github.com/bitcoin/bitcoin/pull/" + pr.number, 0)} target="_blank">${escape(pr.title)}</a></h1> <div class="clearfix m-b-base"></div> ${validate_component(Field, "Field").$$render($$result, { class: "form-field", name: "coverage" }, {}, {
      default: ({ uniqueId }) => {
        return `<label${add_attribute("for", uniqueId, 0)}><i class="ri-file-code-line"></i> <span class="txt" data-svelte-h="svelte-11qq2o1">Coverage report</span></label> ${validate_component(Select, "Select").$$render(
          $$result,
          {
            id: uniqueId,
            toggle: true,
            items: pr.reports || [],
            labelComponent: CoverageReportSelectOption,
            optionComponent: CoverageReportSelectOption,
            selected: selectedReport
          },
          {
            selected: ($$value) => {
              selectedReport = $$value;
              $$settled = false;
            }
          },
          {}
        )}`;
      }
    })}</div> <div class="clearfix m-b-base"></div> ${fetching ? `<div class="alert alert-info" style="text-align: center" data-svelte-h="svelte-1ipxl7k"><i class="ri-information-line"></i> Loading coverage report...</div>` : `${!report ? `<div class="alert alert-info" style="text-align: center" data-svelte-h="svelte-s5uck0"><i class="ri-information-line"></i> No coverage data available.</div>` : ``} ${report && report.status === "pending" ? `<div class="alert alert-warning" style="text-align: center" data-svelte-h="svelte-1gkji5f"><i class="ri-information-line"></i> Coverage report is currently
                    being generated, please come back later.</div>` : ``} ${report && report.status === "failure" ? `<div class="alert alert-danger" style="text-align: center" data-svelte-h="svelte-jkuf39"><i class="ri-information-line"></i> An error occured while generating
                    the coverage report.</div>` : ``} ${report && report.status === "success" ? `${report?.coverage ? `<div class="flex cov-container flex-justify-between flex-align-start bg-grey svelte-1nzmi39"><div class="cov-col svelte-1nzmi39">${validate_component(Coverage, "Coverage").$$render(
      $$result,
      {
        name: "Uncovered new code",
        description: "Lines of code added in this pull request that are not covered by tests.",
        coverage: report.coverage.uncovered_new_code,
        icon: "ri-alert-line",
        color: "txt-danger"
      },
      {},
      {}
    )}</div> <div class="cov-col svelte-1nzmi39">${validate_component(Coverage, "Coverage").$$render(
      $$result,
      {
        name: "Covered new code",
        description: "Lines of code added in this pull request that are covered by tests.",
        coverage: report.coverage.gained_coverage_new_code,
        icon: "ri-check-line",
        color: "txt-success"
      },
      {},
      {}
    )}</div></div> <div class="flex cov-container flex-justify-between flex-align-start svelte-1nzmi39"><div class="cov-col svelte-1nzmi39">${validate_component(Coverage, "Coverage").$$render(
      $$result,
      {
        name: "Lost baseline coverage",
        description: "Lines of code that were covered by tests in master but are not covered anymore in this pull request.",
        coverage: report.coverage.lost_baseline_coverage,
        icon: "ri-alert-line",
        color: "txt-danger"
      },
      {},
      {}
    )}</div> <div class="cov-col svelte-1nzmi39">${validate_component(Coverage, "Coverage").$$render(
      $$result,
      {
        name: "Gained baseline coverage",
        description: "Lines of code that were not covered by tests in master but are covered in this pull request.",
        coverage: report.coverage.gained_baseline_coverage,
        icon: "ri-check-line",
        color: "txt-success"
      },
      {},
      {}
    )}</div></div> <div class="flex cov-container flex-justify-between flex-align-start bg-grey svelte-1nzmi39"><div class="cov-col svelte-1nzmi39">${validate_component(Coverage, "Coverage").$$render(
      $$result,
      {
        name: "Uncovered included code",
        description: "Lines of code that were not executed in master but are executed in this pull request and are not covered by tests.",
        coverage: report.coverage.uncovered_included_code,
        icon: "ri-alert-line",
        color: "txt-danger"
      },
      {},
      {}
    )}</div> <div class="cov-col svelte-1nzmi39">${validate_component(Coverage, "Coverage").$$render(
      $$result,
      {
        name: "Covered included code",
        description: "Lines of code that were not executed in master but are executed in this pull request and are covered by tests.",
        coverage: report.coverage.gained_coverage_included_code,
        icon: "ri-check-line",
        color: "txt-success"
      },
      {},
      {}
    )}</div></div>` : ``} ${sonarcloud ? `<div class="cov-col">${validate_component(Sonarcloud, "Sonarcloud").$$render($$result, { report, issues: sonarcloud.issues }, {}, {})}</div>` : ``} <div class="clearfix m-b-base"></div> <div class="cov-col full-width">${validate_component(Benchmarks, "Benchmarks").$$render($$result, { report }, {}, {})}</div>` : ``}`} <div class="clearfix m-b-base"></div></main> </div>`;
  } while (!$$settled);
  $$unsubscribe_page();
  return $$rendered;
});
export {
  Page as default
};

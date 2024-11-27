import { s as subscribe } from "../../chunks/utils.js";
import { c as create_ssr_component, e as each, a as escape, s as setContext, v as validate_component } from "../../chunks/ssr.js";
import { w as writable } from "../../chunks/index2.js";
import { p as page } from "../../chunks/stores.js";
const toasts = writable([]);
const Toasts = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let $toasts, $$unsubscribe_toasts;
  $$unsubscribe_toasts = subscribe(toasts, (value) => $toasts = value);
  $$unsubscribe_toasts();
  return `<div class="toasts-wrapper">${each($toasts, (toast) => {
    return `<div class="${[
      "alert txt-break",
      (toast.type == "info" ? "alert-info" : "") + " " + (toast.type == "success" ? "alert-success" : "") + " " + (toast.type == "error" ? "alert-danger" : "") + " " + (toast.type == "warning" ? "alert-warning" : "")
    ].join(" ").trim()}"><div class="icon">${toast.type === "info" ? `<i class="ri-information-line"></i>` : `${toast.type === "success" ? `<i class="ri-checkbox-circle-line"></i>` : `${toast.type === "warning" ? `<i class="ri-error-warning-line"></i>` : `<i class="ri-alert-line"></i>`}`}`}</div> <div class="content">${escape(toast.message)}</div> <button type="button" class="close" data-svelte-h="svelte-1l96818"><i class="ri-close-line"></i></button> </div>`;
  })}</div>`;
});
const main = "";
const Layout = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let $page, $$unsubscribe_page;
  $$unsubscribe_page = subscribe(page, (value) => $page = value);
  let { data } = $$props;
  let { user } = data;
  setContext("user", user);
  if ($$props.data === void 0 && $$bindings.data && data !== void 0)
    $$bindings.data(data);
  $$unsubscribe_page();
  return `<div class="app-layout"><aside class="app-sidebar"><a data-sveltekit-preload-code="eager" href="/" class="logo logo-sm" data-svelte-h="svelte-18lmo9e"><img src="${escape("/", true) + "images/logo.png"}" alt="Bitcoin Core Coverage" width="40" height="40"></a> <nav data-sveltekit-preload-data="hover" class="main-menu"><a href="/bitcoin/bitcoin/pulls" class="${["menu-item", $page.url.pathname.includes("/pulls") ? "active" : ""].join(" ").trim()}" aria-label="Pull requests" data-svelte-h="svelte-1pdbwl7"><img style="text-align: center" height="25" alt="pr" src="/icons/svg/pull-request.svg"></a> <a data-sveltekit-preload-code="eager" href="/tests" class="${["menu-item", $page.url.pathname.includes("/tests") ? "active" : ""].join(" ").trim()}" aria-label="Tests" data-svelte-h="svelte-1qeeqq9"><img style="text-align: center" height="25" alt="jobs" src="/icons/svg/ci.svg"></a> <a data-sveltekit-preload-code="eager" href="/benchmarks" class="${[
    "menu-item",
    $page.url.pathname.includes("/benchmarks") ? "active" : ""
  ].join(" ").trim()}" aria-label="Benchmarks" data-svelte-h="svelte-1po5aug"><img style="text-align: center" height="25" alt="jobs" src="/icons/svg/stats.svg"></a> <a data-sveltekit-preload-code="eager" href="/jobs" class="${["menu-item", $page.url.pathname.includes("/jobs") ? "active" : ""].join(" ").trim()}" aria-label="Jobs" data-svelte-h="svelte-la0n40"><img style="text-align: center" height="25" alt="jobs" src="/icons/svg/activity.svg"></a></nav> <a href="https://github.com/corecheck/corecheck" target="_blank" rel="noopener noreferrer" class="menu-item" data-svelte-h="svelte-jo8710"><i style="font-size: 50px; margin-bottom: 10px" class="ri-github-fill"></i></a></aside> <div class="app-body">${slots.default ? slots.default({}) : ``} ${validate_component(Toasts, "Toasts").$$render($$result, {}, {}, {})}</div></div>`;
});
export {
  Layout as default
};

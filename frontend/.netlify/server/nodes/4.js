import * as server from '../entries/pages/_owner_/_repo_/pulls/_number_/_page.server.ts.js';

export const index = 4;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/_owner_/_repo_/pulls/_number_/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/[owner]/[repo]/pulls/[number]/+page.server.ts";
export const imports = ["_app/immutable/nodes/4.f62235f5.js","_app/immutable/chunks/scheduler.ddef2503.js","_app/immutable/chunks/index.fe14b7b1.js","_app/immutable/chunks/stores.ed24c7f0.js","_app/immutable/chunks/singletons.53fe462a.js","_app/immutable/chunks/tooltip.b8cea26f.js","_app/immutable/chunks/SortHeader.7682286d.js"];
export const stylesheets = ["_app/immutable/assets/4.7a7a2964.css"];
export const fonts = [];

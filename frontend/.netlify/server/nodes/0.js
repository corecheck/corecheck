import * as server from '../entries/pages/_layout.server.ts.js';

export const index = 0;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/_layout.svelte.js')).default;
export { server };
export const server_id = "src/routes/+layout.server.ts";
export const imports = ["_app/immutable/nodes/0.6e8ed3a7.js","_app/immutable/chunks/scheduler.ddef2503.js","_app/immutable/chunks/index.fe14b7b1.js","_app/immutable/chunks/tooltip.b8cea26f.js","_app/immutable/chunks/singletons.53fe462a.js","_app/immutable/chunks/stores.ed24c7f0.js"];
export const stylesheets = ["_app/immutable/assets/0.b2248539.css"];
export const fonts = [];

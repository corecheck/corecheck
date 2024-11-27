

export const index = 2;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/_page.svelte.js')).default;
export const imports = ["_app/immutable/nodes/2.e71e2c63.js","_app/immutable/chunks/scheduler.ddef2503.js","_app/immutable/chunks/index.fe14b7b1.js"];
export const stylesheets = [];
export const fonts = [];

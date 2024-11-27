

export const index = 1;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/fallbacks/error.svelte.js')).default;
export const imports = ["_app/immutable/nodes/1.11fa0980.js","_app/immutable/chunks/scheduler.ddef2503.js","_app/immutable/chunks/index.fe14b7b1.js","_app/immutable/chunks/stores.ed24c7f0.js","_app/immutable/chunks/singletons.53fe462a.js"];
export const stylesheets = [];
export const fonts = [];



export const index = 5;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/benchmarks/_page.svelte.js')).default;
export const imports = ["_app/immutable/nodes/5.04d38816.js","_app/immutable/chunks/scheduler.ddef2503.js","_app/immutable/chunks/index.fe14b7b1.js"];
export const stylesheets = [];
export const fonts = [];

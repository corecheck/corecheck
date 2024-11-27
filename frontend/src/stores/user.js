import { writable } from "svelte/store";

// logged app admin
export const user = writable(null);

export function setUser(u) {
    user.set(u);
}

import { redirect, type Actions } from "@sveltejs/kit";
import { env } from '$env/dynamic/public'
import { setUser } from "../stores/user.js";

export async function load({fetch}) {
    const data = await fetch(`${env.PUBLIC_ENDPOINT}/me`, { 
        withCredentials: true,
        credentials: "include",
    }).then((response) => {
        if (response.status === 200) {
            return response.json();
        }
    });

    setUser(data);

    return {
        user: data,
    }
}

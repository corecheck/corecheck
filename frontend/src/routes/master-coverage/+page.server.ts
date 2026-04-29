import { env } from "$env/dynamic/public";

export async function load({ fetch }) {
    try {
        const response = await fetch(`${env.PUBLIC_ENDPOINT}/master-coverage`);
        if (!response.ok) {
            return { report: null };
        }

        return {
            report: await response.json(),
        };
    } catch (error) {
        console.error(error);
        return { report: null };
    }
}

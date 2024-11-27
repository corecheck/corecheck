import { env } from '$env/dynamic/public'

export async function _fetchPulls(title: string, page: number) {
    return fetch(`${env.PUBLIC_ENDPOINT}/pulls?page=${page}&title=${title}`)
        .then(res => {
            if (res.status === 200) {
                return res.json();
            }
            return null;
        })
        .catch((err) => {
            console.error(err);
        });
}

export async function load({ params }) {
    return {
        pulls: await _fetchPulls("", 1)
    };
}
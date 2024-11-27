import { p as public_env } from "../../../../../chunks/shared-server.js";
async function _fetchPulls(title, page) {
  return fetch(`${public_env.PUBLIC_ENDPOINT}/pulls?page=${page}&title=${title}`).then((res) => {
    if (res.status === 200) {
      return res.json();
    }
    return null;
  }).catch((err) => {
    console.error(err);
  });
}
async function load({ params }) {
  return {
    pulls: await _fetchPulls("", 1)
  };
}
export {
  _fetchPulls,
  load
};

import "../../chunks/index.js";
import { p as public_env } from "../../chunks/shared-server.js";
import { w as writable } from "../../chunks/index2.js";
const user = writable(null);
function setUser(u) {
  user.set(u);
}
async function load({ fetch }) {
  const data = await fetch(`${public_env.PUBLIC_ENDPOINT}/me`, {
    withCredentials: true,
    credentials: "include"
  }).then((response) => {
    if (response.status === 200) {
      return response.json();
    }
  });
  setUser(data);
  return {
    user: data
  };
}
export {
  load
};

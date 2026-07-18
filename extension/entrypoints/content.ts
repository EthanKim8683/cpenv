import { addListener } from "@/rpc/content/add-listener";

export default defineContentScript({
  matches: ["https://*.codeforces.com/*"],
  main() {
    addListener();
  },
});

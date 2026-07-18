import { addListener } from "@/src/message-rpc/add-listener";
import { methods } from "@/src/content/methods";

export default defineContentScript({
  matches: ["https://*.codeforces.com/*"],
  main() {
    addListener(methods);
  },
});

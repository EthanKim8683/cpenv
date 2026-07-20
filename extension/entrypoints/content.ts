import { addMessageRpcListeners } from "@/message-rpc/add-listeners";
import { contentMethods } from "@/content/methods";

export default defineContentScript({
  matches: ["https://*.codeforces.com/*"],
  main() {
    addMessageRpcListeners(contentMethods);
  },
});

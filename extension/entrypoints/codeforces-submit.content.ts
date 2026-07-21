import { submitClient } from "@/lib/server";
import { getTabId } from "./tab-id.content";

export default defineContentScript({
  matches: ["https://*.codeforces.com/contest/*/submit"],
  runAt: "document_start",
  allFrames: true,
  world: "MAIN",
  async main() {
    if (window.parent.frames.length === 0) return;

    window.stop = () => {};

    for await (const response of submitClient.subscribe({
      tabId: getTabId(),
    })) {
      console.log(response);
    }
  },
});

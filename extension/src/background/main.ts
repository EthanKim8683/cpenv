import { isInternal } from "./lib/tabs";
import { handleTab as handleCodeforcesTab } from "./codeforces/handle-tab";

async function handleTab(tab: chrome.tabs.Tab) {
  if (await isInternal(tab.id!)) return;

  const { hostname } = new URL(tab.url!);
  if (!hostname) return;

  if (hostname.includes("codeforces.com")) {
    await handleCodeforcesTab(tab);
  }
}

chrome.tabs.onCreated.addListener((tab) => {
  handleTab(tab);
});

chrome.tabs.onUpdated.addListener((_tabId, _changeInfo, tab) => {
  handleTab(tab);
});

const TAB_ID_STORAGE_KEY = "cpenv-tab-id";

export function getTabId() {
  return sessionStorage.getItem(TAB_ID_STORAGE_KEY)!;
}

export default defineContentScript({
  matches: ["https://*.codeforces.com/*"],
  main() {
    const tabId = crypto.randomUUID();
    sessionStorage.setItem(TAB_ID_STORAGE_KEY, tabId);
  },
});

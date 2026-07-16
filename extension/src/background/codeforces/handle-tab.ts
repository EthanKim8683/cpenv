import { handleProblemTab } from "./handle-problem-tab";
import { handleContestTab } from "./handle-contest-tab";

export async function handleTab(tab: chrome.tabs.Tab) {
  const { pathname } = new URL(tab.url!);
  let match: string[] | null = null;
  if (
    (match = /\/contest\/(\d+)\/problem\/(\w+)/.exec(pathname)) ||
    (match = /\/problemset\/problem\/(\d+)\/(\w+)/.exec(pathname))
  ) {
    handleProblemTab(tab, match[1], match[2]);
  } else if ((match = /\/contest\/(\d+)/.exec(pathname))) {
    handleContestTab(tab, match[1]);
  }
}

import { Mutex } from "async-mutex";

const INTERNAL_TAB_GROUP_TITLE = "cpenv";

const internalTabMutex = new Mutex();

async function getNormalWindowId(): Promise<number> {
  const windows = await chrome.windows.getAll({
    windowTypes: ["normal"],
    populate: false,
  });
  return (windows.find((window) => window.focused) ?? windows[0])!.id!;
}

async function getInternalTabGroupId(): Promise<number | undefined> {
  const groups = await chrome.tabGroups.query({
    windowId: await getNormalWindowId(),
    title: INTERNAL_TAB_GROUP_TITLE,
  });
  return groups[0]?.id!;
}

export function isInternal(tabId: number) {
  return internalTabMutex.runExclusive(async () => {
    const [tab, maybeGroupId] = await Promise.all([
      chrome.tabs.get(tabId),
      getInternalTabGroupId(),
    ]);
    return tab.groupId === maybeGroupId;
  });
}

export async function waitForComplete(
  tabId: number,
  { timeoutMs = 5_000 }: { timeoutMs?: number } = {},
) {
  const { promise, resolve, reject } = Promise.withResolvers<void>();

  const handleUpdated = (
    updatedTabId: number,
    changeInfo: chrome.tabs.OnUpdatedInfo,
  ) => {
    if (updatedTabId !== tabId) return;
    if (changeInfo.status !== "complete") return;
    resolve();
  };

  const handleRemoved = (removedTabId: number) => {
    if (removedTabId !== tabId) return;
    reject(new Error(`Tab with id ${tabId} was removed.`));
  };

  (async () => {
    const tab = await chrome.tabs.get(tabId);
    if (tab.status === "complete") {
      resolve();
    }
  })();

  const handleTimeout = () => {
    reject(new Error(`Tab with id ${tabId} timed out.`));
  };

  chrome.tabs.onUpdated.addListener(handleUpdated);
  chrome.tabs.onRemoved.addListener(handleRemoved);
  const timeout = setTimeout(handleTimeout, timeoutMs);
  await promise;
  chrome.tabs.onUpdated.removeListener(handleUpdated);
  chrome.tabs.onRemoved.removeListener(handleRemoved);
  clearTimeout(timeout);
}

export async function navigate(
  tabId: number,
  url: string,
  {
    timeoutMs = 5_000,
    retries = 5,
  }: { timeoutMs?: number; retries?: number } = {},
) {
  let lastError: unknown;
  for (let i = 0; i < retries; i++) {
    try {
      await chrome.tabs.update(tabId, { url });
      await waitForComplete(tabId, { timeoutMs });
      return;
    } catch (error) {
      lastError = error;
      await chrome.tabs.reload(tabId);
    }
  }
  throw lastError;
}

export async function createInternalTab(url: string) {
  const tab = await internalTabMutex.runExclusive(async () => {
    const [tab, maybeGroupId] = await Promise.all([
      chrome.tabs.create({
        url: "about:blank",
        active: false,
      }),
      getInternalTabGroupId(),
    ]);

    const groupId = await chrome.tabs.group({
      tabIds: [tab.id!],
      groupId: maybeGroupId,
    });

    if (!maybeGroupId) {
      chrome.tabGroups.update(groupId, {
        collapsed: true,
        title: INTERNAL_TAB_GROUP_TITLE,
      });
    }

    return tab;
  });

  await navigate(tab.id!, url);

  return tab;
}

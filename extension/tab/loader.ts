class TimeoutError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "TimeoutError";
  }
}

export type LoaderOptions = {
  timeout?: number;
  retries?: number;
};

export class Loader {
  private timeout: number;
  private retries: number;

  constructor(options: LoaderOptions = {}) {
    this.timeout = options.timeout ?? 5_000;
    this.retries = options.retries ?? 3;
  }

  async loadTab(tabId: number) {
    const attempt = async () => {
      const { promise, resolve, reject } = Promise.withResolvers<void>();

      const onUpdated = async (
        updatedTabId: number,
        changeInfo: Browser.tabs.OnUpdatedInfo,
      ) => {
        if (updatedTabId !== tabId) return;
        if (changeInfo.status !== "complete") return;
        resolve();
      };

      const onRemoved = async (removedTabId: number) => {
        if (removedTabId !== tabId) return;
        reject(new Error(`Tab ${tabId} was removed while loading.`));
      };

      (async () => {
        const tab = await browser.tabs.get(tabId);
        if (tab.status !== "complete") return;
        resolve();
      })();

      const onTimeout = () => {
        reject(new TimeoutError(`Tab ${tabId} timed out while loading.`));
      };

      browser.tabs.onUpdated.addListener(onUpdated);
      browser.tabs.onRemoved.addListener(onRemoved);
      const timeout = setTimeout(onTimeout, this.timeout);
      await promise;
      browser.tabs.onUpdated.removeListener(onUpdated);
      browser.tabs.onRemoved.removeListener(onRemoved);
      clearTimeout(timeout);
    };
    for (let i = 0; i <= this.retries; i++) {
      try {
        return await attempt();
      } catch (error) {
        if (!(error instanceof TimeoutError)) throw error;
      }
    }
    throw new Error(
      `Tab ${tabId} failed to load after ${this.retries + 1} attempts.`,
    );
  }
}

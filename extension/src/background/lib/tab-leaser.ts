export class TabLeaser {
  leaseTab(
    urls: string[],
    callback: (tab: chrome.tabs.Tab) => PromiseLike<void> | void,
  ): void {
    const tab = chrome.tabs.create({ url: urls[0] });
    callback(tab);
  }
}

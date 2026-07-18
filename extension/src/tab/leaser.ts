import { Hider } from "./hider";
import assert from "@/lib/assert";

export type LeaserOptions = {
  hider?: Hider;
};

export class Leaser {
  private hider: Hider;

  constructor(options: LeaserOptions = {}) {
    this.hider = options.hider ?? new Hider();
  }

  async createAndLeaseTab<T>(
    callback: (tabId: number) => PromiseLike<T> | T,
  ): Promise<T> {
    const tab = await browser.tabs.create({
      url: "about:blank",
      active: false,
    });
    assert(tab.id);
    const result = await callback(tab.id);
    await browser.tabs.remove(tab.id);
    return result;
  }

  async createAndLeaseHiddenTab<T>(
    callback: (tabId: number) => PromiseLike<T> | T,
  ): Promise<T> {
    const tabId = await this.hider.createHiddenTab();
    assert(tabId);
    const result = await callback(tabId);
    await browser.tabs.remove(tabId);
    return result;
  }
}

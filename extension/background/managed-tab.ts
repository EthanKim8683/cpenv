import { TabIdGetter } from "./ports/tab-id-getter";

export class ManagedTab {
  async getTabId(): Promise<number> {
    return 0;
  }
}

({}) as ManagedTab satisfies TabIdGetter;

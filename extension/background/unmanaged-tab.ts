import { TabIdGetter } from "./ports/tab-id-getter";

export class UnmanagedTab {
  async getTabId(): Promise<number> {
    return 0;
  }
}

({}) as UnmanagedTab satisfies TabIdGetter;

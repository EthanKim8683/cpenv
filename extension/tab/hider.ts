import { Mutex } from "async-mutex";
import assert from "@/lib/assert";

const HIDDEN_TAB_GROUP_TITLE = "cpenv";

export class Hider {
  private mutex = new Mutex();

  private async getHiddenTabGroupId() {
    const groups = await browser.tabGroups.query({
      title: HIDDEN_TAB_GROUP_TITLE,
    });
    return groups[0]?.id;
  }

  private async internalIsTabHidden(tabId: number) {
    const [tab, groupId] = await Promise.all([
      browser.tabs.get(tabId),
      this.getHiddenTabGroupId(),
    ]);
    if (!tab || !groupId) return false;
    return tab.groupId === groupId;
  }

  private async internalHideTab(tabId: number) {
    const maybeGroupId = await this.getHiddenTabGroupId();
    const groupId = await browser.tabs.group({
      tabIds: [tabId],
      groupId: maybeGroupId,
    });
    if (maybeGroupId) return;

    await Promise.all([
      browser.tabGroups.update(groupId, {
        title: HIDDEN_TAB_GROUP_TITLE,
      }),
      browser.tabGroups.move(groupId, {
        index: 0,
      }),
    ]);
  }

  private async internalUnhideTab(tabId: number) {
    if (!(await this.internalIsTabHidden(tabId))) return;
    await browser.tabs.ungroup(tabId);
  }

  private async internalCreateHiddenTab() {
    const tab = await browser.tabs.create({
      url: "about:blank",
      active: false,
    });
    assert(tab.id);
    await this.internalHideTab(tab.id);
    return tab.id;
  }

  async isTabHidden(tabId: number) {
    return this.mutex.runExclusive(async () => {
      return this.internalIsTabHidden(tabId);
    });
  }

  async hideTab(tabId: number) {
    return this.mutex.runExclusive(async () => {
      return this.internalHideTab(tabId);
    });
  }

  async unhideTab(tabId: number) {
    return this.mutex.runExclusive(async () => {
      return this.internalUnhideTab(tabId);
    });
  }

  async createHiddenTab() {
    return this.mutex.runExclusive(async () => {
      return this.internalCreateHiddenTab();
    });
  }
}

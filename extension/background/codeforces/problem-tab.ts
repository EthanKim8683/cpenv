import { TabIdGetter } from "../ports/tab-id-getter";

export type CodeforcesProblemTabOptions = {
  tabIdGetter: TabIdGetter;
};

export class CodeforcesProblemTab {
  private tabIdGetter: TabIdGetter;

  constructor(options: CodeforcesProblemTabOptions) {
    this.tabIdGetter = options.tabIdGetter;
  }
}

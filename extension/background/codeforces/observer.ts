import { Leaser } from "@/tab/leaser";
import { Loader } from "@/tab/loader";
import { ContestResolver } from "./contest-resolver";
import { ProblemResolver } from "./problem-resolver";
import assert from "@/lib/assert";

export type ObserverOptions = {
  leaser?: Leaser;
  loader?: Loader;
  contestResolver?: ContestResolver;
  problemResolver?: ProblemResolver;
};

export class Observer {
  private contestResolver: ContestResolver;
  private problemResolver: ProblemResolver;

  constructor(options: ObserverOptions = {}) {
    this.contestResolver = options.contestResolver ?? new ContestResolver();
    this.problemResolver = options.problemResolver ?? new ProblemResolver();
  }

  async addListeners() {
    const resolve = async (tabId: number) => {
      const tab = await browser.tabs.get(tabId);
      assert(tab.url);
      const u = new URL(tab.url);
      let match: string[] | null = null;
      if (
        (match = /\/contest\/(\d+)\/problem\/(\w+)$/.exec(u.pathname)) ||
        (match = /\/problemset\/problem\/(\d+)\/(\w+)$/.exec(u.pathname))
      ) {
        const [, contestId, problemIndex] = match;
        const problem = await this.problemResolver.resolveProblem(
          tabId,
          contestId,
          problemIndex,
        );
        console.log(problem);
      } else if ((match = /\/contest\/(\d+)$/.exec(u.pathname))) {
        const [, contestId] = match;
        const contest = await this.contestResolver.resolveContest(
          tabId,
          contestId,
        );
        console.log(contest);
      }
    };

    const onCreated = async (tab: Browser.tabs.Tab) => {
      await resolve(tab.id!);
    };

    const onUpdated = async (tabId: number) => {
      await resolve(tabId);
    };

    browser.tabs.onCreated.addListener(onCreated);
    browser.tabs.onUpdated.addListener(onUpdated);
    return () => {
      browser.tabs.onCreated.removeListener(onCreated);
      browser.tabs.onUpdated.removeListener(onUpdated);
    };
  }
}

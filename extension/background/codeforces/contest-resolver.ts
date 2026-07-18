import { Contest, ContestSchema } from "@/gen/contest/v1/contest_pb";
import { createClient } from "@/rpc/content/create-client";
import { Leaser } from "@/tab/leaser";
import { Loader } from "@/tab/loader";
import { ProblemResolver } from "./problem-resolver";
import { create } from "@bufbuild/protobuf";

export type ContestResolverOptions = {
  leaser?: Leaser;
  loader?: Loader;
  problemResolver?: ProblemResolver;
};

export class ContestResolver {
  private leaser: Leaser;
  private loader: Loader;
  private problemResolver: ProblemResolver;

  constructor(options: ContestResolverOptions = {}) {
    this.leaser = options.leaser ?? new Leaser();
    this.loader = options.loader ?? new Loader();
    this.problemResolver = options.problemResolver ?? new ProblemResolver();
  }

  async resolveContest(tabId: number, contestId: string): Promise<Contest> {
    await this.loader.loadTab(tabId);

    const client = createClient(tabId);
    const data = await client.scrapeCodeforcesContest();

    const promises = [];
    for (const problemIndex of data.problemIndices) {
      promises.push(
        this.leaser.createAndLeaseHiddenTab(async (tabId) => {
          await browser.tabs.update(tabId, {
            url: `https://codeforces.com/contest/${contestId}/problem/${problemIndex}`,
          });
          await this.loader.loadTab(tabId);

          const problem = await this.problemResolver.resolveProblem(
            tabId,
            contestId,
            problemIndex,
          );
          return [problemIndex, problem];
        }),
      );
    }

    return create(ContestSchema, {
      id: ["codeforces", contestId],
      problems: Object.fromEntries(await Promise.all(promises)),
    });
  }
}

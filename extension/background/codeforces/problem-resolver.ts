import { Problem, ProblemSchema } from "@/gen/problem/v1/problem_pb";
import { createClient } from "@/rpc/content/create-client";
import { Loader } from "@/tab/loader";
import { create } from "@bufbuild/protobuf";

export type ProblemResolverOptions = {
  loader?: Loader;
};

export class ProblemResolver {
  private loader: Loader;

  constructor(options: ProblemResolverOptions = {}) {
    this.loader = options.loader ?? new Loader();
  }

  async resolveProblem(
    tabId: number,
    contestId: string,
    problemIndex: string,
  ): Promise<Problem> {
    await this.loader.loadTab(tabId);
    const client = createClient(tabId);
    const data = await client.scrapeCodeforcesProblem();

    return create(ProblemSchema, {
      id: [contestId, problemIndex],
      ...data,
    });
  }
}

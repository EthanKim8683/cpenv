import { contentServer } from "../../content/server";
import { create } from "@bufbuild/protobuf";
import { createInternalTab } from "../lib/tabs";
import { observationClient } from "../lib/api";
import { ReportContestRequestSchema } from "../../gen/observation/v1/observation_service_pb";

export async function handleContestTab(
  tab: chrome.tabs.Tab,
  contestId: string,
) {
  const client = contentServer.createClient(tab.id!);
  const data = await client.scrapeCodeforcesContest();
  const promises = [];
  for (const problemIndex of data.problemIndices) {
    promises.push(
      (async () => {
        const problemTab = await createInternalTab(
          `https://codeforces.com/contest/${contestId}/problem/${problemIndex}`,
        );
        const problemClient = contentServer.createClient(problemTab.id!);
        const problem = await problemClient.scrapeCodeforcesProblem();
        await problemClient.close();
        return [problemIndex, problem] as const;
      })(),
    );
  }
  const problems = Object.fromEntries(await Promise.all(promises));
  const contest = create(ReportContestRequestSchema, {
    contest: {
      id: contestId,
      problems,
    },
  });
  chrome.storage.local.set({
    [`${import.meta.url}~${tab.id!}`]: contest,
  });
}

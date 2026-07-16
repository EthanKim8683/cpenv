import { contentServer } from "../../content/server";
import { create } from "@bufbuild/protobuf";
import { observationClient } from "../lib/api";
import { ReportProblemRequestSchema } from "../../gen/observation/v1/observation_service_pb";

export async function handleProblemTab(
  tab: chrome.tabs.Tab,
  contestId: string,
  problemIndex: string,
) {
  const client = contentServer.createClient(tab.id!);
  const data = await client.scrapeCodeforcesProblem();
  const problem = create(ReportProblemRequestSchema, {
    problem: {
      io: data.io,
      type: data.type,
      samples: data.samples,
    },
  });
  chrome.storage.local.set({
    [`${import.meta.url}~${tab.id!}`]: problem,
  });
}

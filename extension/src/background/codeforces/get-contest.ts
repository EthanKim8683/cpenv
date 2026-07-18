import { Contest } from "@/gen/contest/v1/contest_pb";
import { client } from "../client";

export async function getContest({
  contestId,
}: {
  contestId: string;
}): Promise<Contest> {
  const problemIndex = "A";
  client.call(`/codeforces/problem/${contestId}/${problemIndex}`);
}

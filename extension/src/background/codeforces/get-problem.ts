import { Problem } from "@/gen/problem/v1/problem_pb";

export async function getProblem({
  contestId,
  problemIndex,
}: {
  contestId: string;
  problemIndex: string;
}): Promise<Problem> {}

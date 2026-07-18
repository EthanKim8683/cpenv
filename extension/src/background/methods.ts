import { getContest as getCodeforcesContest } from "./codeforces/get-contest";
import { getProblem as getCodeforcesProblem } from "./codeforces/get-problem";

export const methods = {
  "/codeforces/contest/:contestId": getCodeforcesContest,
  "/codeforces/problem/:contestId/:problemIndex": getCodeforcesProblem,
};

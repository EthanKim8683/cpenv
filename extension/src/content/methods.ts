import { scrapeContest as scrapeCodeforcesContest } from "./codeforces/scrape-contest";
import { scrapeProblem as scrapeCodeforcesProblem } from "./codeforces/scrape-problem";

export const methods = {
  scrapeCodeforcesContest,
  scrapeCodeforcesProblem,
};

export type Methods = typeof methods;

import { scrapeContest as scrapeCodeforcesContest } from "@/content/codeforces/scrape-contest";
import { scrapeProblem as scrapeCodeforcesProblem } from "@/content/codeforces/scrape-problem";

export const methods = {
  scrapeCodeforcesContest,
  scrapeCodeforcesProblem,
};

export type Methods = typeof methods;

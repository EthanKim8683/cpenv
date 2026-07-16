import { Server } from "../../lib/server";
import { close } from "./close";
import { scrapeContest as scrapeCodeforcesContest } from "./codeforces/scrape-contest";
import { scrapeProblem as scrapeCodeforcesProblem } from "./codeforces/scrape-problem";

export const contentServer = new Server("content", {
  close,
  scrapeCodeforcesContest,
  scrapeCodeforcesProblem,
} as const);

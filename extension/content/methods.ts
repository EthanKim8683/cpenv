import { scrapeCodeforcesProblem } from "./codeforces/scrape-problem";

export const contentMethods = {
  scrapeCodeforcesProblem,
};

export type ContentMethods = typeof contentMethods;

function scrapeProblemIndices() {
  return Array.from(
    document
      .querySelectorAll("table.problems tbody tr td:first-child")
      .values()
      .map((problem) => (problem as HTMLElement).textContent.trim()),
  );
}

export function scrapeContest() {
  return {
    problemIndices: scrapeProblemIndices(),
  };
}

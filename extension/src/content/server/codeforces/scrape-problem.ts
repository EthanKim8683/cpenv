import { ProblemIo, ProblemType } from "../../../gen/problem/v1/problem_pb";

function scrapeSectionTitles() {
  return Array.from(
    document
      .querySelectorAll("div.section-title, span > span.tex-font-style-bf")
      .values()
      .map((element) => {
        let title = (element as HTMLElement).textContent;
        title = title.replace(/\s+/g, " ");
        title = title.trim();
        title = title.toLowerCase();
        return title;
      }),
  );
}

function scrapeProblemType() {
  const sectionTitles = scrapeSectionTitles();
  if (
    sectionTitles.includes("first run") &&
    sectionTitles.includes("second run")
  ) {
    return ProblemType.RUN_TWICE;
  } else if (sectionTitles.includes("interaction")) {
    return ProblemType.INTERACTIVE;
  } else {
    return ProblemType.BATCH;
  }
}

function scrapeSamples() {
  return Array.from(
    document
      .querySelectorAll("div.sample-test")
      .values()
      .map((sample, index) => {
        let element: HTMLElement | null = null;

        let input = "";
        if ((element = sample.querySelector("div.input pre"))) {
          input = element.innerText;
        } else {
          console.warn(`No input found for sample ${index + 1}.`);
        }

        let output = "";
        if ((element = sample.querySelector("div.output pre"))) {
          output = element.innerText;
        } else {
          console.warn(`No output found for sample ${index + 1}.`);
        }

        return { input, output };
      })
      .filter((sample) => sample !== null),
  );
}

export function scrapeProblem() {
  const io = ProblemIo.STDIO; // TODO: Explore non-stdio problems
  const type = scrapeProblemType();
  const samples = scrapeSamples();

  return {
    io,
    type,
    samples,
  };
}

import { FocusProblemRequestSchema } from "@/gen/extension/v1/extension_service_pb";
import { ProblemType } from "@/gen/problem/v1/problem_pb";
import { serverClient } from "@/lib/server";
import { create } from "@bufbuild/protobuf";
import { getTabId } from "./tab-id.content";

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
    return ProblemType.STDIO_RUN_TWICE;
  } else if (sectionTitles.includes("interaction")) {
    return ProblemType.STDIO_INTERACTIVE;
  } else {
    return ProblemType.STDIO_BATCH;
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

function injectSubmitIframe(contestId: string) {
  const iframe = document.createElement("iframe");
  iframe.src = `https://codeforces.com/contest/${contestId}/submit`;
  // iframe.style.display = "none";
  document.body.appendChild(iframe);
}

export default defineContentScript({
  matches: [
    "https://codeforces.com/contest/*/problem/*",
    "https://codeforces.com/problemset/problem/*/*",
    "https://*.codeforces.com/contest/*/problem/*",
    "https://*.codeforces.com/problemset/problem/*/*",
  ],
  main() {
    const url = window.location.href;
    const match =
      /codeforces\.com\/contest\/(\d+)\/problem\/(\w+)/.exec(url) ??
      /codeforces\.com\/problemset\/problem\/(\d+)\/(\w+)/.exec(url);
    const [, contestId, problemIndex] = match!;

    injectSubmitIframe(contestId);

    const type = scrapeProblemType();
    const samples = scrapeSamples();

    serverClient.focusProblem(
      create(FocusProblemRequestSchema, {
        tabId: getTabId(),
        problem: {
          id: `codeforces-${contestId}-${problemIndex}`,
          type,
          samples,
        },
      }),
    );
  },
});

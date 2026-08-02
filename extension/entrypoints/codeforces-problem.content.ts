import { ProblemType } from "@/gen/problem/v1/problem_pb";
import { createProblemMain } from "@/lib/server";

function getProblemId() {
  const url = window.location.href;
  const match =
    /codeforces\.com\/contest\/(\d+)\/problem\/(\w+)/.exec(url) ??
    /codeforces\.com\/problemset\/problem\/(\d+)\/(\w+)/.exec(url);
  const [, contestId, problemIndex] = match!;
  return `codeforces-${contestId}-${problemIndex}`.toLowerCase();
}

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
  const inputs = Array.from(
    document
      .querySelectorAll("div.sample-test div.input pre")
      .values()
      .map((element) => (element as HTMLPreElement).innerText),
  );

  const outputs = Array.from(
    document
      .querySelectorAll("div.sample-test div.output pre")
      .values()
      .map((element) => (element as HTMLPreElement).innerText),
  );

  const samples = [];
  for (let i = 0; i < Math.max(inputs.length, outputs.length); i++) {
    const input = inputs[i] ?? "";
    if (!input) {
      console.warn(`No input found for sample ${i + 1}.`);
    }

    const output = outputs[i] ?? "";
    if (!output) {
      console.warn(`No output found for sample ${i + 1}.`);
    }

    samples.push({ input, output });
  }
  return samples;
}

function scrapeProblem() {
  const id = getProblemId();
  const type = scrapeProblemType();
  const samples = scrapeSamples();
  return { id, type, samples };
}

function getProgramTypeId(sourceFile: File) {
  const selectElement = document.querySelector(
    "select[name='programTypeId']",
  ) as HTMLSelectElement;

  let patterns: RegExp[] = [];
  const extension = sourceFile.name.split(".").pop()?.toLowerCase();
  switch (extension) {
    case "cpp":
      patterns.push(/\bg\+\+\b/i);
      break;
    case "py":
      patterns.push(/\bpypy\b/i);
      break;
    default:
      throw new Error(`Unsupported file extension: ${extension}.`);
  }

  let value = -1;
  for (const option of selectElement.options) {
    if (!patterns.some((pattern) => pattern.test(option.text))) continue;
    value = Math.max(value, parseInt(option.value));
  }
  if (value === -1) {
    console.warn(`No program type found for file ${sourceFile.name}.`);
    return (
      selectElement.querySelector("option[selected]") as HTMLOptionElement
    ).value;
  }

  return value.toString();
}

async function submit(sourceFile: File) {
  const formElement = document.querySelector(".submitForm") as HTMLFormElement;
  const formData = new FormData(formElement);

  formData.set("programTypeId", getProgramTypeId(sourceFile));
  formData.set("sourceFile", sourceFile);

  const url = window.location.href;
  const response = await fetch(url, {
    method: "POST",
    body: formData,
    credentials: "include",
  });

  if (!response.ok) {
    throw new Error(`Failed to submit problem: ${response.statusText}.`);
  }
}

export default defineContentScript({
  matches: [
    "https://codeforces.com/contest/*/problem/*",
    "https://codeforces.com/problemset/problem/*/*",
    "https://*.codeforces.com/contest/*/problem/*",
    "https://*.codeforces.com/problemset/problem/*/*",
  ],
  main: createProblemMain({
    getProblemId,
    scrapeProblem,
    submit,
  }),
});

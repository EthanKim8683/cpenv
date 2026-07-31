import { Focus, FocusSchema } from "@/gen/focus/v1/focus_pb";
import { SetFocusRequestSchema } from "@/gen/focus/v1/focus_service_pb";
import { ProblemType } from "@/gen/problem/v1/problem_pb";
import {
  SubscribeRequestSchema,
  CallbackRequestSchema,
} from "@/gen/submit/v1/submit_service_pb";
import { focusClient, submitClient } from "@/lib/server";
import { create, MessageInitShape } from "@bufbuild/protobuf";

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
  throw new Error("Not implemented");

  const formElement = document.querySelector(".submitForm") as HTMLFormElement;
  const formData = new FormData(formElement);

  formData.set("programTypeId", getProgramTypeId(sourceFile));
  formData.set("sourceFile", sourceFile);

  const response = await fetch(window.location.href, {
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
  async main() {
    const url = window.location.href;
    const match =
      /codeforces\.com\/contest\/(\d+)\/problem\/(\w+)/.exec(url) ??
      /codeforces\.com\/problemset\/problem\/(\d+)\/(\w+)/.exec(url);
    const [, contestId, problemIndex] = match!;
    const id = `codeforces-${contestId}-${problemIndex}`.toLowerCase();

    let focus: MessageInitShape<typeof FocusSchema> = {};
    try {
      const type = scrapeProblemType();
      const samples = scrapeSamples();

      focus = {
        problem: {
          id,
          type,
          samples,
        },
      };
    } catch (caughtError: unknown) {
      focus = {
        error: String(caughtError),
      };
    }

    const handleVisibilityChange = async () => {
      if (document.visibilityState !== "visible") return;

      await focusClient.setFocus(
        create(SetFocusRequestSchema, {
          focus,
        }),
      );
    };

    handleVisibilityChange();
    document.addEventListener("visibilitychange", handleVisibilityChange);

    for await (const {
      callbackId,
      content,
      fileName,
    } of submitClient.subscribe(
      create(SubscribeRequestSchema, {
        problemId: id,
      }),
    )) {
      let error: string | undefined;
      try {
        const sourceFile = new File([new Uint8Array(content)], fileName);
        await submit(sourceFile);
      } catch (caughtError: unknown) {
        error = String(caughtError);
      }

      await submitClient.callback(
        create(CallbackRequestSchema, {
          callbackId,
          error,
        }),
      );
    }
  },
});

import { createConnectTransport } from "@connectrpc/connect-web";
import { createClient } from "@connectrpc/connect";
import {
  FocusService,
  SetFocusRequestSchema,
} from "@/gen/focus/v1/focus_service_pb";
import {
  CallbackRequestSchema,
  SubmitService,
  SubscribeRequestSchema,
} from "@/gen/submit/v1/submit_service_pb";
import { create, MessageInitShape } from "@bufbuild/protobuf";
import { ProblemSchema } from "@/gen/problem/v1/problem_pb";
import { FocusSchema } from "@/gen/focus/v1/focus_pb";
import retry from "async-retry";

const retryOpts = {
  forever: true,
  factor: 2,
  minTimeout: 500,
  maxTimeout: 60_000,
  randomize: false,
  onRetry: (error: unknown, attempt: number) => {
    const timeout = Math.min(
      retryOpts.minTimeout * retryOpts.factor ** (attempt - 1),
      retryOpts.maxTimeout,
    );
    console.warn(`Disconnected from server; retrying in ${timeout}ms`, error);
  },
};

const transport = createConnectTransport({
  baseUrl: `http://localhost:${import.meta.env.WXT_PORT}`,
});

const focusClient = createClient(FocusService, transport);
const submitClient = createClient(SubmitService, transport);

type MaybePromise<T> = T | PromiseLike<T>;

export function createProblemMain({
  getProblemId,
  scrapeProblem,
  submit,
}: {
  getProblemId: () => string;
  scrapeProblem: () => MaybePromise<MessageInitShape<typeof ProblemSchema>>;
  submit: (file: File) => MaybePromise<void>;
}) {
  return async () => {
    let focus: MessageInitShape<typeof FocusSchema> | undefined;
    try {
      focus = {
        problem: await scrapeProblem(),
      };
    } catch (caughtError: unknown) {
      focus = {
        error: String(caughtError),
      };
    }

    let currentEventId = 0;
    const handleVisibilityChange = async () => {
      if (document.visibilityState !== "visible") return;

      currentEventId++;
      const eventId = currentEventId;

      await retry<void, void>(async (bail) => {
        if (currentEventId !== eventId) {
          bail();
          return;
        }

        await focusClient.setFocus(
          create(SetFocusRequestSchema, {
            focus,
          }),
        );
      }, retryOpts).catch();
    };
    handleVisibilityChange();
    document.addEventListener("visibilitychange", handleVisibilityChange);

    await retry(async () => {
      for await (const {
        callbackId,
        fileName,
        content,
      } of submitClient.subscribe(
        create(SubscribeRequestSchema, {
          problemId: getProblemId(),
        }),
      )) {
        let error: string | undefined;
        try {
          await submit(new File([new Uint8Array(content)], fileName));
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

      throw new Error("Disconnected gracefully.");
    }, retryOpts);
  };
}

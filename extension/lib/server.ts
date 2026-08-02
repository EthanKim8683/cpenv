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

const initialReconnectDelayMs = 500;
const reconnectDelayGrowthFactor = 1.5;
const maxReconnectDelayMs = 60_000;

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

    let reconnectDelay = initialReconnectDelayMs;
    while (true) {
      try {
        for await (const {
          callbackId,
          fileName,
          content,
        } of submitClient.subscribe(
          create(SubscribeRequestSchema, {
            problemId: getProblemId(),
          }),
        )) {
          reconnectDelay = initialReconnectDelayMs;

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
      } catch (error: unknown) {
        console.warn(
          `Disconnected from server; reconnecting in ${reconnectDelay}ms`,
          error,
        );
      }

      await new Promise((resolve) => setTimeout(resolve, reconnectDelay));
      reconnectDelay = Math.min(
        reconnectDelay * reconnectDelayGrowthFactor,
        maxReconnectDelayMs,
      );
    }
  };
}

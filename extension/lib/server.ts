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

type MaybePromise<T> = T | PromiseLike<T>;

type GroupRetrierOptions = {
  minDelay: number;
  maxDelay: number;
  growthFactor: number;
};

type GroupRetrierFn = (stop: () => void) => MaybePromise<void>;

class GroupRetrier {
  private fns = new Set<GroupRetrierFn>();
  private retryLoop: Promise<void> | undefined;

  constructor(private readonly opts: GroupRetrierOptions) {}

  async add(fn: GroupRetrierFn) {
    try {
      await fn(() => {});
    } catch (error: unknown) {
      this.fns.add(fn);
      console.warn("Adding to retry loop:", error);

      if (!this.retryLoop) {
        this.retryLoop = (async () => {
          let delay = this.opts.minDelay;
          while (true) {
            const results = await Promise.allSettled(
              Array.from(this.fns).map(async (fn) => {
                const stop = () => this.fns.delete(fn);
                await fn(stop);
                stop();
              }),
            );
            if (this.fns.size === 0) break;

            delay = Math.min(
              delay * this.opts.growthFactor,
              this.opts.maxDelay,
            );
            console.warn(
              `Retrying in ${delay}ms:`,
              ...results
                .filter((result) => result.status === "rejected")
                .map((result) => result.reason),
            );

            await new Promise((resolve) => setTimeout(resolve, delay));
          }
          this.retryLoop = undefined;
        })();
      }
    }
  }
}

const transport = createConnectTransport({
  baseUrl: `http://localhost:${import.meta.env.WXT_PORT}`,
});

const focusClient = createClient(FocusService, transport);
const submitClient = createClient(SubmitService, transport);

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

    const retrier = new GroupRetrier({
      minDelay: 500,
      maxDelay: 60_000,
      growthFactor: 2,
    });

    let currentEventId = 0;
    const handleVisibilityChange = async () => {
      if (document.visibilityState !== "visible") return;

      currentEventId++;
      const eventId = currentEventId;

      retrier.add(async (stop) => {
        if (currentEventId !== eventId) {
          stop();
          return;
        }

        await focusClient.setFocus(
          create(SetFocusRequestSchema, {
            focus,
          }),
        );
      });
    };
    handleVisibilityChange();
    document.addEventListener("visibilitychange", handleVisibilityChange);

    retrier.add(async () => {
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
    });
  };
}

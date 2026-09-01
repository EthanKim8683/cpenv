import { createConnectTransport } from "@connectrpc/connect-web";
import { createClient } from "@connectrpc/connect";
import {
  ActiveProblemService,
  SaveRequestSchema,
} from "@/gen/active_problem/v1/active_problem_service_pb";
import {
  SubmitService,
  ClaimRequestSchema,
  ReplyRequestSchema,
} from "@/gen/submit/v1/submit_service_pb";
import { create, MessageInitShape } from "@bufbuild/protobuf";
import { ProblemSchema } from "@/gen/problem/v1/problem_pb";
import { ActiveProblemSchema } from "@/gen/active_problem/v1/active_problem_pb";

const PORT = import.meta.env.WXT_PORT ?? "8683";

type MaybePromise<T> = T | PromiseLike<T>;

const transport = createConnectTransport({
  baseUrl: `http://localhost:${PORT}`,
});

const activeProblemClient = createClient(ActiveProblemService, transport);
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
    let activeProblem: MessageInitShape<typeof ActiveProblemSchema> | undefined;
    try {
      activeProblem = { problem: await scrapeProblem() };
    } catch (caughtError: unknown) {
      if (typeof caughtError !== "string") {
        activeProblem = { error: `uncaught error: ${caughtError}` };
      } else {
        activeProblem = { error: caughtError };
      }
    }

    const handleVisibilityChange = async () => {
      if (document.visibilityState !== "visible") return;
      await activeProblemClient.save(
        create(SaveRequestSchema, { activeProblem }),
      );
    };
    handleVisibilityChange();
    document.addEventListener("visibilitychange", handleVisibilityChange);

    while (true) {
      const { content, fileName, replyId } = await submitClient.claim(
        create(ClaimRequestSchema, { problemId: getProblemId() }),
      );

      let error: string | undefined;
      try {
        await submit(new File([new Uint8Array(content)], fileName));
      } catch (caughtError: unknown) {
        if (typeof caughtError !== "string") {
          error = `uncaught error: ${caughtError}`;
        } else {
          error = caughtError;
        }
      }

      await submitClient.reply(create(ReplyRequestSchema, { replyId, error }));
    }
  };
}

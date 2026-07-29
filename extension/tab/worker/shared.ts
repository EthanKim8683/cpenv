import { z } from "zod";

export const WORKER_NAMESPACE = "cpenv-worker";

export const WorkerMessageSchema = z.object({
  ns: z.literal(WORKER_NAMESPACE),
  req: z.unknown(),
});

export type WorkerMessage = z.infer<typeof WorkerMessageSchema>;

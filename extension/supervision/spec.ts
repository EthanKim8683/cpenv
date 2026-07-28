import type { Worker } from "./worker/types";

export type Spec = {
  canHandle: (req: unknown) => boolean;
  createWorker: (req: unknown) => Promise<Worker>;
  idleTimeoutMs: number;
};

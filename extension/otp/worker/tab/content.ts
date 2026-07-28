import { WorkerMessageSchema } from "./shared";
import type { Worker } from "../types";

export abstract class TabWorker {
  abstract requestHandler(
    req: unknown,
  ): Promise<(() => Promise<unknown>) | null>;

  kill(): void {
    window.close();
  }
}

({}) as TabWorker satisfies Worker;

export function exposeTabWorker(worker: TabWorker) {
  browser.runtime.onMessage.addListener((message) => {
    const { success, data } = WorkerMessageSchema.safeParse(message);
    if (!success) return;

    return worker.requestHandler(data.req);
  });
}

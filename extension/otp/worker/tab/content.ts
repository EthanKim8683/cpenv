import { WorkerMessageSchema } from "./shared";

export interface TabWorker {
  requestHandler(req: unknown): Promise<(() => Promise<unknown>) | null>;
}

export function exposeTabWorker(worker: TabWorker) {
  browser.runtime.onMessage.addListener((message) => {
    const { success, data } = WorkerMessageSchema.safeParse(message);
    if (!success) return;

    return worker.requestHandler(data.req);
  });
}

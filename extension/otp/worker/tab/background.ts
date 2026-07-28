import { WORKER_NAMESPACE, type WorkerMessage } from "./shared";
import type { Worker } from "../types";

export class TabWorkerProxy {
  constructor(private readonly tabId: number) {}

  requestHandler(req: unknown): Promise<(() => Promise<unknown>) | null> {
    return browser.tabs.sendMessage(this.tabId, {
      ns: WORKER_NAMESPACE,
      req,
    } satisfies WorkerMessage);
  }

  kill(): void {
    browser.tabs.remove(this.tabId);
  }
}

({}) as TabWorkerProxy satisfies Worker;

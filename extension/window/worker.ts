import type { Spec } from "@/supervision/spec";
import type { Worker } from "@/supervision/worker";

export const windowSpec: Spec = {
  canHandle: (req: unknown) => true,
  createWorker: (req: unknown) =>
    Promise.resolve(new WindowWorker(browser.windows.WINDOW_ID_NONE)),
  idleTimeoutMs: 0,
};

export class WindowWorker {
  constructor(private readonly windowId: number) {}

  requestHandler(req: unknown): Promise<(() => Promise<unknown>) | null> {
    browser.windows.get(this.windowId);
  }

  kill(): void {
    browser.windows.remove(this.windowId);
  }
}

({}) as WindowWorker satisfies Worker;

import type { Worker } from "../worker/types";
import type { Spec } from "../spec";
import assert from "assert";
import { BusMessageSchema } from "./shared";

export type HandleOptions = {
  acceptTimeoutMs?: number;
  handleTimeoutMs?: number;
};

export class Bus {
  private readonly specs: Spec[] = [];
  private readonly workers: Map<Worker, NodeJS.Timeout> = new Map();

  registerSpec(spec: Spec): void {
    this.specs.push(spec);
  }

  async handle(
    req: unknown,
    { acceptTimeoutMs = 5_000, handleTimeoutMs = 5_000 }: HandleOptions = {},
  ): Promise<unknown> {
    const promises: Promise<() => Promise<unknown>>[] = [];
    for (const [worker, timeout] of this.workers) {
      promises.push(
        (async () => {
          const handler = await worker.requestHandler(req);
          if (!handler) throw undefined;
          if (!this.workers.has(worker)) throw undefined;

          timeout.refresh();
          return handler;
        })(),
      );
    }
    const handler = await Promise.race([
      new Promise<never>((_resolve, reject) =>
        setTimeout(() => reject(), acceptTimeoutMs),
      ),
      Promise.any(promises),
    ]).catch(async () => {
      for (const spec of this.specs) {
        if (!spec.canHandle(req)) continue;

        const worker = await spec.createWorker(req);
        const handler = await worker.requestHandler(req);
        assert(handler);

        this.workers.set(
          worker,
          setTimeout(() => {
            worker.kill();
            this.workers.delete(worker);
          }, spec.idleTimeoutMs),
        );

        return handler;
      }

      throw new Error("Unhandled request");
    });

    return await Promise.race([
      new Promise<never>((_resolve, reject) => {
        setTimeout(
          () => reject(new Error("Handler timed out")),
          handleTimeoutMs,
        );
      }),
      handler(),
    ]);
  }
}

export function exposeBus(bus: Bus) {
  browser.runtime.onMessage.addListener(
    (message): Promise<unknown> | undefined => {
      const { success, data } = BusMessageSchema.safeParse(message);
      if (!success) return;

      return bus.handle(data.req, data.opts);
    },
  );
}

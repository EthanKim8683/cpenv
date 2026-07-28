import type { HandleOptions } from "./background";
import { BUS_NAMESPACE, type BusMessage } from "./shared";

export class BusProxy {
  handle(req: unknown, opts?: HandleOptions): Promise<unknown> {
    return browser.runtime.sendMessage({
      ns: BUS_NAMESPACE,
      req,
      opts,
    } satisfies BusMessage);
  }
}

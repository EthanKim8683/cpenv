import { Methods } from "./methods";
import { CONTENT_RPC_NAMESPACE, Message } from "./message";

export function createClient(tabId: number) {
  return new Proxy(
    {} as {
      [K in keyof Methods]: (
        ...args: Parameters<Methods[K]>
      ) => Promise<Awaited<ReturnType<Methods[K]>>>;
    },
    {
      get(_target, p) {
        return (...args: any[]) => {
          return browser.tabs.sendMessage(tabId, {
            namespace: CONTENT_RPC_NAMESPACE,
            runtimeId: browser.runtime.id,
            method: p as string,
            args,
          } satisfies Message);
        };
      },
    },
  );
}

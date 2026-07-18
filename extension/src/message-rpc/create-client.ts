import { CONTENT_RPC_NAMESPACE, Message } from "./message";
import { MethodMap, Client } from "./types";

export function createClient<T extends MethodMap>(tabId: number) {
  return new Proxy({} as Client<T>, {
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
  });
}

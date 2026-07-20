import { messageSchema } from "./message";
import { MethodMap } from "./types";

export function addMessageRpcListeners(methods: MethodMap) {
  const onMessage = async (message: any) => {
    const { success, data } = messageSchema.safeParse(message);
    if (!success) return;

    const method = methods[data.method];
    if (!method) {
      throw new Error(`Method '${data.method}' not found`);
    }
    return method(...data.args);
  };

  browser.runtime.onMessage.addListener(onMessage);
  return () => {
    browser.runtime.onMessage.removeListener(onMessage);
  };
}

import { z } from "zod";

const MESSAGE_TYPE = `cpenv(${chrome.runtime.id})`;

const messageSchema = z.object({
  type: z.literal(MESSAGE_TYPE),
  runtimeId: z.string(),
  serverId: z.string(),
  method: z.string(),
  args: z.array(z.any()),
});

type Message = z.infer<typeof messageSchema>;

export type Handlers = {
  [key: string]: (...args: any[]) => any;
};

export class Server<T extends Handlers> {
  constructor(
    private id: string,
    private server: T,
  ) {}

  addListener() {
    chrome.runtime.onMessage.addListener(async (message) => {
      const { success, data } = messageSchema.safeParse(message);
      if (!success) return;

      if (data.runtimeId !== chrome.runtime.id) return;
      if (data.serverId !== this.id) return;
      return this.server[data.method](...data.args);
    });
  }

  createClient(tabId: number): {
    [K in keyof T]: (
      ...args: Parameters<T[K]>
    ) => PromiseLike<ReturnType<T[K]>>;
  } {
    return Object.fromEntries(
      Object.entries(this.server).map(([key, _value]) => [
        key,
        (...args) =>
          chrome.tabs.sendMessage(tabId, {
            type: MESSAGE_TYPE,
            runtimeId: chrome.runtime.id,
            serverId: this.id,
            method: key,
            args,
          } satisfies Message),
      ]),
    ) as T;
  }
}

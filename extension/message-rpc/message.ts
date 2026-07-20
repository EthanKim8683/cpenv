import { z } from "zod";

export const CONTENT_RPC_NAMESPACE = "cpenv-content-rpc";

export const messageSchema = z.object({
  namespace: z.literal(CONTENT_RPC_NAMESPACE),
  runtimeId: z.literal(browser.runtime.id),
  method: z.string(),
  args: z.array(z.any()),
});

export type Message = z.infer<typeof messageSchema>;

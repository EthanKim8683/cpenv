import { z } from "zod";
import type { HandleOptions } from "./background";

export const BUS_NAMESPACE = "cpenv-bus";

export const BusMessageSchema = z.object({
  ns: z.literal(BUS_NAMESPACE),
  req: z.unknown(),
  opts: (
    z.object({
      acceptTimeoutMs: z.number().optional(),
      handleTimeoutMs: z.number().optional(),
    }) satisfies z.ZodType<HandleOptions>
  ).optional(),
});

export type BusMessage = z.infer<typeof BusMessageSchema>;

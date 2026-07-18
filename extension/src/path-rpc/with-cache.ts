import { Client, MethodMap } from "./types";
import { Simplify } from "type-fest";

export function withCache<
  T extends Client<Methods, Options>,
  Methods extends MethodMap,
  Options extends {},
>(
  client: T,
): Client<Methods, Simplify<Options & { revalidateCache?: boolean }>> {
  const cache = new Map<string, any>();
  return {
    // @ts-ignore If there are required option properties, TypeScript will
    // complain, preventing the default value from being used.
    call: (path, options = {}) => {
      const cached = cache.get(path);
      if (cached && !options.revalidateCache) return cached;
      const result = client.call(path, options);
      cache.set(path, result);
      return result;
    },
  };
}

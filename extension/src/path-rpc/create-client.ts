import { Client, NarrowMethods } from "./types";

export function createClient<T extends NarrowMethods<T>>(
  methods: T,
): Client<T, {}> {
  const entries: [RegExp, (...args: any[]) => any][] = [];
  for (const path in methods) {
    const pattern = new RegExp(
      `^${path
        .split("/")
        .map((part) => {
          if (part.startsWith(":")) {
            return `(?<${part.slice(1)}>[^/]*)`;
          }
          // @ts-ignore Older versions of TypeScript may not be aware of
          // RegExp.escape.
          return RegExp.escape(part);
        })
        .join("/")}$`,
    );
    entries.push([pattern, methods[path]]);
  }

  return {
    call: (path, _options) => {
      for (const [pattern, method] of entries) {
        const match = path.match(pattern);
        if (!match) continue;
        return method(match.groups ?? {});
      }
      throw new Error(`Unhandled path: ${path}`);
    },
  };
}

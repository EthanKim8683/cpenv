import type { RemovePrefix, Simplify, Split } from "type-fest";

export type MethodMap = Record<string, (...args: any[]) => any>;

type _GetParams<Segments extends string[]> = Segments extends [
  infer First extends string,
  ...infer Rest extends string[],
]
  ? First extends `:${infer Param extends string}`
    ? { [K in Param]: string } & _GetParams<Rest>
    : _GetParams<Rest>
  : {};

type GetParams<Path extends string> = Simplify<_GetParams<Split<Path, "/">>>;

export type NarrowMethods<Methods extends MethodMap> = {
  [K in Extract<keyof Methods, string>]: (
    params: GetParams<K>,
  ) => ReturnType<Methods[K]>;
};

type _GetPattern<Segments extends string[]> = Segments extends [
  infer First extends string,
  ...infer Rest extends string[],
]
  ? First extends `:${string}`
    ? `/${string}${_GetPattern<Rest>}`
    : `/${First}${_GetPattern<Rest>}`
  : "";

type GetPattern<Path extends string> = RemovePrefix<
  _GetPattern<Split<Path, "/">>,
  "/"
>;

export interface Client<Methods extends MethodMap, Options extends {}> {
  call<
    Path extends {
      [K in keyof Methods]: K extends string ? GetPattern<K> : never;
    }[keyof Methods],
  >(
    path: Path,
    options?: Options,
  ): {
    [K in keyof Methods]: K extends string
      ? Path extends GetPattern<K>
        ? ReturnType<Methods[K]>
        : GetPattern<K> extends Path
          ? ReturnType<Methods[K]>
          : never
      : never;
  }[keyof Methods];
}

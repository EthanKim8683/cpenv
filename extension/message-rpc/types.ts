export type MethodMap = Record<string, (...args: any[]) => any>;

export type Client<T extends MethodMap> = {
  [K in keyof T]: (
    ...args: Parameters<T[K]>
  ) => Promise<Awaited<ReturnType<T[K]>>>;
};

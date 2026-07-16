import { Mutex } from "async-mutex";

export class CacheUpdater<T> {
  private cache = new Map<string, T>();
  private mutex = new Mutex();

  async updateCache(
    key: string,
    callback: (cached: T | null) => PromiseLike<T> | T,
  ): Promise<T> {
    return this.mutex.runExclusive(async () => {
      const cachedValue = this.cache.get(key) ?? null;
      const value = await callback(cachedValue);
      this.cache.set(key, value);
      return value;
    });
  }
}

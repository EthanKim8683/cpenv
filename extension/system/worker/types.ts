export interface Worker {
  requestHandler(req: unknown): Promise<(() => Promise<unknown>) | null>;
  kill(): void;
}

export type Spec = {
  canHandle: (req: unknown) => boolean;
  createWorker: (req: unknown) => Promise<Worker>;
  idleTimeoutMs: number;
};

export interface Worker {
  requestHandler(req: unknown): Promise<(() => Promise<unknown>) | null>;
  kill(): void;
}

// when we make requests to the bus, the bus forwards requests to sidecars
//
// sidecars respond with a unique id if their worker can handle the request (or
// don't respond at all if they can't)
//
// the bus will see the first response (if any) and send the responding sidecar
// the go-ahead to fulfill the request
//
// the sidecar will tell the worker to fulfill the request and return the result
// to the bus

// or how about this:
//
// we solidify the background-content divide:
//
// background:
// - bus: same
// - sidecars: same but responds with tab id rather than response id
//
// content:
// - workers: same
//
// and then everything else is not worker-y

import { Bus, exposeBus } from "@/system/bus/background";

export default defineBackground({
  main() {
    const bus = new Bus();
    bus.registerSpec({
      canHandle: (req) => req.type === "codeforces-problem",
      createWorker: (req) => new CodeforcesProblemWorker(req),
      idleTimeoutMs: 10000,
    });
    exposeBus(bus);
  },
});

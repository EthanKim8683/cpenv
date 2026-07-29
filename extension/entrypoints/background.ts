import { bus } from "@/supervision/bus/background";
import { windowSpec } from "@/window/worker";

export default defineBackground({
  main() {
    bus.registerSpec(windowSpec);
  },
});

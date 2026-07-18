import { Observer } from "@/background/codeforces/observer";
import { Loader } from "@/tab/loader";
import { Hider } from "@/tab/hider";
import { Leaser } from "@/tab/leaser";
import { ContestResolver } from "@/background/codeforces/contest-resolver";
import { ProblemResolver } from "@/background/codeforces/problem-resolver";

export default defineBackground(() => {
  const loader = new Loader({});
  const hider = new Hider();
  const leaser = new Leaser({
    hider: hider,
  });
  const problemResolver = new ProblemResolver({
    loader: loader,
  });
  const contestResolver = new ContestResolver({
    leaser: leaser,
    loader: loader,
    problemResolver: problemResolver,
  });
  const observer = new Observer({
    contestResolver: contestResolver,
    problemResolver: problemResolver,
  });
  observer.addListeners();
});

export default defineContentScript({
  matches: ["https://*.codeforces.com/contest/*/submit"],
  runAt: "document_start",
  allFrames: true,
  world: "MAIN",
  main() {
    if (window.parent.frames.length === 0) return;

    window.stop = () => {};
  },
});

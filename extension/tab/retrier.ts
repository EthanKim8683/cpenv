// // watches tabs
// //
// // we could use observers but to be honest i don't think it makes that much of a
// // difference and is more complicated
// //
// // uses backoff

// // tbh rewrite

// import { Loader, LoadOptions } from "./loader";
// import retry, { Options } from "async-retry";

// class CheckError extends Error {
//   constructor(message: string) {
//     super(message);
//     this.name = "CheckError";
//   }
// }

// export class Watcher {
//   private loader: Loader;

//   constructor(private tabId: number) {
//     this.loader = new Loader(tabId);
//   }

//   async watch(
//     check: () => Promise<boolean>,
//     {
//       retry: { onRetry, ...rest },
//       load: loadOptions,
//     }: {
//       retry: Options;
//       load: LoadOptions;
//     },
//   ) {
//     await retry(
//       async () => {
//         await this.loader.load(loadOptions);
//         if (await check()) return;
//         throw new CheckError("Check failed.");
//       },
//       {
//         onRetry: (e, attempt) => {
//           if (!(e instanceof CheckError)) throw e;
//           onRetry?.(e, attempt);
//           browser.tabs.reload(this.tabId);
//         },
//         ...rest,
//       },
//     );
//   }
// }

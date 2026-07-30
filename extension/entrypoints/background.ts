export default defineBackground({
  async main() {
    // when we receive a "watch submissions" request, we open the right
    // submissions page and the page will watch itself and send us updates. if
    // it's gone quiet before finishing, reload it. this'll be a nice catch-all
    // for submissions pages that do or don't watch
  },
});

import { defineConfig } from "wxt";

// See https://wxt.dev/api/config.html
export default defineConfig({
  webExt: {
    disabled: true,
  },
  manifest: {
    host_permissions: ["http://localhost/*", "http://127.0.0.1/*"],
  },
});

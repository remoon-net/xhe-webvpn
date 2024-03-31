import { defineConfig } from "vite";

export default defineConfig({
  build: {
    lib: {
      entry: ".",
      name: "vpn",
      formats: ["umd"],
    },
  },
  resolve: {}
});

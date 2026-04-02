import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";
import { nodePolyfills } from "vite-plugin-node-polyfills";

export default defineConfig({
  plugins: [
    react(),
    nodePolyfills({ include: ["buffer", "process"], globals: { Buffer: true, process: true } }),
  ],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  define: {
    "process.env": {},
    global: "globalThis",
  },
  optimizeDeps: {
    include: ["@solana/web3.js", "@coral-xyz/anchor", "buffer"],
    esbuildOptions: {
      target: "esnext",
    },
  },
  build: {
    target: "esnext",
    rollupOptions: {
      output: {
        manualChunks: {
          solana: ["@solana/web3.js", "@coral-xyz/anchor"],
          wallet: [
            "@solana/wallet-adapter-react",
            "@solana/wallet-adapter-react-ui",
          ],
        },
      },
    },
  },
});

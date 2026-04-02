import React, { useCallback } from "react";
import ReactDOM from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter } from "react-router-dom";
import {
  ConnectionProvider,
  WalletProvider,
} from "@solana/wallet-adapter-react";
import { WalletModalProvider } from "@solana/wallet-adapter-react-ui";
import { WalletError, WalletNotReadyError } from "@solana/wallet-adapter-base";
import { PhantomWalletAdapter } from "@solana/wallet-adapter-phantom";
import { SolflareWalletAdapter } from "@solana/wallet-adapter-solflare";
import { clusterApiUrl } from "@solana/web3.js";
import App from "./App";
import "@solana/wallet-adapter-react-ui/styles.css";
import "./index.css";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 2,
    },
  },
});

// Standard wallet adapter auto-detects installed wallets (Phantom, Solflare, Backpack).
// Specific adapters below are included so users see wallet options + install links
// in the modal even when not yet installed.
const wallets = [
  new PhantomWalletAdapter(),
  new SolflareWalletAdapter(),
];

const endpoint = import.meta.env.VITE_RPC_URL ?? clusterApiUrl("devnet");

function Root() {
  const onWalletError = useCallback((error: WalletError) => {
    if (error instanceof WalletNotReadyError) {
      // Extension not installed — silently ignore autoConnect attempts on page load.
      // If user explicitly clicked a wallet in the modal, the modal already
      // shows the install link; no need for an alert.
      console.warn("[wallet] not ready:", error.message);
    } else {
      console.error("[wallet]", error.name, error.message);
    }
  }, []);

  return (
    <ConnectionProvider endpoint={endpoint}>
      <WalletProvider wallets={wallets} autoConnect onError={onWalletError}>
        <WalletModalProvider>
          <QueryClientProvider client={queryClient}>
            <BrowserRouter>
              <App />
            </BrowserRouter>
          </QueryClientProvider>
        </WalletModalProvider>
      </WalletProvider>
    </ConnectionProvider>
  );
}

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <Root />
  </React.StrictMode>
);

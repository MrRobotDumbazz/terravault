import { useEffect, useRef } from "react";
import { useWallet } from "@solana/wallet-adapter-react";
import { useWalletStore } from "@/store/wallet";

const API_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

/**
 * After a wallet connects, automatically:
 * 1. Request a challenge nonce from the backend
 * 2. Ask the wallet to sign the message
 * 3. Verify the signature → receive JWT with role
 * 4. Store token + role in Zustand
 *
 * Clears auth when wallet disconnects.
 */
export function useAuth() {
  const { publicKey, signMessage, connected, disconnecting } = useWallet();
  const { setAuth, clearAuth, token } = useWalletStore();
  const authInProgress = useRef(false);

  useEffect(() => {
    // Wallet disconnected → clear auth
    if (!connected || disconnecting) {
      clearAuth();
      authInProgress.current = false;
      return;
    }

    // Already authenticated or in progress
    if (!publicKey || !signMessage || token || authInProgress.current) return;

    authInProgress.current = true;

    (async () => {
      try {
        const wallet = publicKey.toBase58();

        // 1. Get challenge
        const challengeRes = await fetch(`${API_URL}/api/v1/auth/challenge`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ wallet }),
        });
        if (!challengeRes.ok) throw new Error("Failed to get challenge");
        const { message, nonce: _nonce } = await challengeRes.json();

        // 2. Sign message
        const encoded = new TextEncoder().encode(message);
        const signature = await signMessage(encoded);

        // 3. Verify signature → JWT
        // Default role = investor; admin is assigned server-side via ADMIN_WALLETS
        const verifyRes = await fetch(`${API_URL}/api/v1/auth/verify`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            wallet,
            signature: btoa(String.fromCharCode(...signature)),
            role: "investor",
          }),
        });
        if (!verifyRes.ok) throw new Error("Signature verification failed");
        const { token: jwt, role } = await verifyRes.json();

        setAuth(jwt, role);
      } catch (err) {
        console.error("[useAuth] auth flow failed:", err);
      } finally {
        authInProgress.current = false;
      }
    })();
  }, [connected, disconnecting, publicKey, signMessage, token]);
}

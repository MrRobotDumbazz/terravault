import { useQuery } from "@tanstack/react-query";
import { useWallet, useConnection } from "@solana/wallet-adapter-react";
import { PublicKey } from "@solana/web3.js";
import { getInvestorPositionPDA } from "@/lib/pda";
import { API_BASE_URL } from "@/lib/constants";

/** Fetch investor's positions from Go API */
export function useMyPositions() {
  const { publicKey } = useWallet();

  return useQuery({
    queryKey: ["positions", publicKey?.toBase58()],
    queryFn: async () => {
      const res = await fetch(`${API_BASE_URL}/v1/me/positions`, {
        headers: {
          // JWT injected by auth middleware
          Authorization: `Bearer ${localStorage.getItem("tv_token") ?? ""}`,
        },
      });
      if (!res.ok) throw new Error("Failed to fetch positions");
      return res.json();
    },
    enabled: !!publicKey,
  });
}

/** Fetch unclaimed distributions for investor */
export function useUnclaimedDistributions(projectId?: string) {
  const { publicKey } = useWallet();

  return useQuery({
    queryKey: ["distributions", projectId, publicKey?.toBase58()],
    queryFn: async () => {
      const url = projectId
        ? `${API_BASE_URL}/v1/projects/${projectId}/distributions`
        : `${API_BASE_URL}/v1/me/positions`;
      const res = await fetch(url, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem("tv_token") ?? ""}`,
        },
      });
      if (!res.ok) throw new Error("Failed to fetch distributions");
      return res.json();
    },
    enabled: !!publicKey,
  });
}

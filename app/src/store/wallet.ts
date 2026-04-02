import { create } from "zustand";

type Role = "investor" | "developer" | "admin" | null;

interface WalletStore {
  role: Role;
  token: string | null;
  kycVerified: boolean;
  kycTimestamp: number | null;
  setAuth: (token: string, role: Role) => void;
  clearAuth: () => void;
  setKycVerified: (verified: boolean, timestamp?: number) => void;
}

export const useWalletStore = create<WalletStore>((set) => ({
  role: null,
  token: null,
  kycVerified: false,
  kycTimestamp: null,
  setAuth: (token, role) => set({ token, role }),
  clearAuth: () => set({ token: null, role: null }),
  setKycVerified: (verified, timestamp) =>
    set({ kycVerified: verified, kycTimestamp: timestamp ?? null }),
}));

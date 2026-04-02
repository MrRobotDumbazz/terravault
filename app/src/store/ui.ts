import { create } from "zustand";

interface Toast {
  id: string;
  type: "success" | "error" | "info" | "warning";
  title: string;
  description?: string;
}

interface UIStore {
  toasts: Toast[];
  addToast: (toast: Omit<Toast, "id">) => void;
  removeToast: (id: string) => void;
  investModalOpen: boolean;
  setInvestModalOpen: (open: boolean) => void;
  activeProjectId: string | null;
  setActiveProjectId: (id: string | null) => void;
}

export const useUIStore = create<UIStore>((set) => ({
  toasts: [],
  addToast: (toast) =>
    set((state) => ({
      toasts: [...state.toasts, { ...toast, id: crypto.randomUUID() }],
    })),
  removeToast: (id) =>
    set((state) => ({
      toasts: state.toasts.filter((t) => t.id !== id),
    })),
  investModalOpen: false,
  setInvestModalOpen: (open) => set({ investModalOpen: open }),
  activeProjectId: null,
  setActiveProjectId: (id) => set({ activeProjectId: id }),
}));

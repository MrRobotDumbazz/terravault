import { useQuery } from "@tanstack/react-query";
import { useConnection } from "@solana/wallet-adapter-react";
import { PublicKey } from "@solana/web3.js";
import { API_BASE_URL } from "@/lib/constants";
import type { ApiProject, Project, ProjectMetadata } from "@/types/project";

/** Fetch project list from Go API (fast, indexed) */
export function useProjects(filters?: {
  type?: string;
  status?: string;
  page?: number;
}) {
  return useQuery({
    queryKey: ["projects", filters],
    queryFn: async () => {
      const params = new URLSearchParams();
      if (filters?.type) params.set("type", filters.type);
      if (filters?.status) params.set("status", filters.status);
      if (filters?.page) params.set("page", String(filters.page));
      const res = await fetch(`${API_BASE_URL}/v1/projects?${params}`);
      if (!res.ok) throw new Error("Failed to fetch projects");
      return res.json() as Promise<{ projects: ApiProject[]; total: number }>;
    },
  });
}

/** Fetch single project detail from Go API */
export function useProjectDetail(id: string) {
  return useQuery({
    queryKey: ["project", id],
    queryFn: async () => {
      const res = await fetch(`${API_BASE_URL}/v1/projects/${id}`);
      if (!res.ok) throw new Error("Project not found");
      return res.json() as Promise<ApiProject & { metadata: ProjectMetadata }>;
    },
    enabled: !!id,
  });
}

/** Fetch project milestones */
export function useProjectMilestones(projectId: string) {
  return useQuery({
    queryKey: ["milestones", projectId],
    queryFn: async () => {
      const res = await fetch(
        `${API_BASE_URL}/v1/projects/${projectId}/milestones`
      );
      if (!res.ok) throw new Error("Failed to fetch milestones");
      return res.json();
    },
    enabled: !!projectId,
  });
}

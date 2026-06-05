import { useCallback, useEffect, useState } from "react";

export interface Usecase {
  id: string;
  name: string;
  description: string;
  steps: number;
}

export function useUsecases() {
  const [usecases, setUsecases] = useState<Usecase[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("/api/usecases")
      .then((res) => res.json())
      .then((data: { usecases: Usecase[] }) => {
        setUsecases(data.usecases);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const run = useCallback(async (usecaseId: string, sessionId: string) => {
    const res = await fetch(`/api/usecases/${encodeURIComponent(usecaseId)}/run`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ session_id: sessionId }),
    });

    if (!res.ok) {
      throw new Error(`Failed to start usecase (${res.status})`);
    }

    return (await res.json()) as { usecase_id: string; session_id: string };
  }, []);

  return { usecases, loading, run };
}

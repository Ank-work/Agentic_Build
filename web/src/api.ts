export type BuildStatus = "queued" | "running" | "succeeded" | "failed";

export interface Build {
  id: number;
  name: string;
  branch: string;
  status: BuildStatus;
  createdAt: string;
}

async function json<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const body = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(body.error ?? `Request failed with ${res.status}`);
  }
  return res.json() as Promise<T>;
}

export function listBuilds(): Promise<Build[]> {
  return fetch("/api/builds").then((r) => json<Build[]>(r));
}

export function createBuild(input: { name: string; branch: string }): Promise<Build> {
  return fetch("/api/builds", {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(input),
  }).then((r) => json<Build>(r));
}

export function advanceBuild(id: number): Promise<Build> {
  return fetch(`/api/builds/${id}/advance`, { method: "POST" }).then((r) =>
    json<Build>(r),
  );
}

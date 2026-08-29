const STATUS_FLOW = ["queued", "running", "succeeded"];

/**
 * In-memory store for build pipeline records. Kept intentionally simple so the
 * app runs with zero external infrastructure; swap for a real database later.
 */
export function createStore() {
  let nextId = 1;
  /** @type {Array<{id:number,name:string,branch:string,status:string,createdAt:string}>} */
  const builds = [];

  function create({ name, branch }) {
    const build = {
      id: nextId++,
      name: String(name).trim(),
      branch: String(branch || "main").trim(),
      status: "queued",
      createdAt: new Date().toISOString(),
    };
    builds.unshift(build);
    return build;
  }

  function list() {
    return builds;
  }

  function get(id) {
    return builds.find((b) => b.id === Number(id));
  }

  function advance(id) {
    const build = get(id);
    if (!build) return undefined;
    const idx = STATUS_FLOW.indexOf(build.status);
    if (idx >= 0 && idx < STATUS_FLOW.length - 1) {
      build.status = STATUS_FLOW[idx + 1];
    }
    return build;
  }

  function seed() {
    create({ name: "api-service", branch: "main" });
    create({ name: "web-frontend", branch: "feature/ui" });
    builds[builds.length - 1].status = "succeeded";
  }

  return { create, list, get, advance, seed };
}

export { STATUS_FLOW };

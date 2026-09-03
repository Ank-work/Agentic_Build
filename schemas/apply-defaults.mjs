/**
 * Executor defaults from the job protocol (§6.1 invariants):
 * missing network / network.mode → { mode: "none" }.
 * Apply defaults, then validate; this is not a schema relaxation.
 */
export function applyDefaults(job) {
  const out = JSON.parse(JSON.stringify(job ?? {}));
  if (out.spec == null || typeof out.spec !== "object" || Array.isArray(out.spec)) {
    out.spec = {};
  }
  const net = out.spec.network;
  if (net == null || typeof net !== "object" || Array.isArray(net)) {
    out.spec.network = { mode: "none" };
  } else if (net.mode == null || net.mode === "") {
    out.spec.network = { ...net, mode: "none" };
  }
  return out;
}

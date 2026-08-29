import { useEffect, useState } from "react";
import { advanceBuild, createBuild, listBuilds, type Build } from "./api";

export default function App() {
  const [builds, setBuilds] = useState<Build[]>([]);
  const [name, setName] = useState("");
  const [branch, setBranch] = useState("main");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  async function refresh() {
    try {
      setBuilds(await listBuilds());
      setError(null);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    refresh();
  }, []);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;
    try {
      await createBuild({ name: name.trim(), branch: branch.trim() || "main" });
      setName("");
      setBranch("main");
      await refresh();
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function onAdvance(id: number) {
    try {
      await advanceBuild(id);
      await refresh();
    } catch (err) {
      setError((err as Error).message);
    }
  }

  const succeeded = builds.filter((b) => b.status === "succeeded").length;
  const running = builds.filter((b) => b.status === "running").length;

  return (
    <div className="app">
      <header className="hero">
        <div className="logo">▲</div>
        <div>
          <h1>Agentic Build</h1>
          <p>Track and drive your agent build pipelines.</p>
        </div>
      </header>

      <section className="stats">
        <div className="stat">
          <span className="stat-value">{builds.length}</span>
          <span className="stat-label">Total</span>
        </div>
        <div className="stat">
          <span className="stat-value">{running}</span>
          <span className="stat-label">Running</span>
        </div>
        <div className="stat">
          <span className="stat-value">{succeeded}</span>
          <span className="stat-label">Succeeded</span>
        </div>
      </section>

      <form className="new-build" onSubmit={onSubmit}>
        <input
          aria-label="Build name"
          placeholder="Service or app name"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <input
          aria-label="Branch"
          placeholder="Branch"
          value={branch}
          onChange={(e) => setBranch(e.target.value)}
        />
        <button type="submit">Queue build</button>
      </form>

      {error && <p className="error">{error}</p>}

      <section className="builds">
        {loading ? (
          <p className="empty">Loading…</p>
        ) : builds.length === 0 ? (
          <p className="empty">No builds yet. Queue your first one above.</p>
        ) : (
          builds.map((b) => (
            <article key={b.id} className="build-card">
              <div className="build-main">
                <h3>{b.name}</h3>
                <span className="branch">{b.branch}</span>
              </div>
              <div className="build-side">
                <span className={`badge badge-${b.status}`}>{b.status}</span>
                {b.status !== "succeeded" && b.status !== "failed" && (
                  <button className="advance" onClick={() => onAdvance(b.id)}>
                    Advance →
                  </button>
                )}
              </div>
            </article>
          ))
        )}
      </section>
    </div>
  );
}

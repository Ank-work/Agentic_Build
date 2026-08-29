import assert from "node:assert/strict";
import test from "node:test";
import { createApp } from "./app.js";
import { createStore } from "./store.js";

function startServer() {
  const app = createApp({ store: createStore(), serveClient: false });
  const server = app.listen(0);
  const { port } = server.address();
  const base = `http://127.0.0.1:${port}`;
  return { server, base };
}

test("health endpoint reports ok", async () => {
  const { server, base } = startServer();
  try {
    const res = await fetch(`${base}/api/health`);
    assert.equal(res.status, 200);
    const body = await res.json();
    assert.equal(body.status, "ok");
  } finally {
    server.close();
  }
});

test("builds start empty and can be created", async () => {
  const { server, base } = startServer();
  try {
    let res = await fetch(`${base}/api/builds`);
    assert.deepEqual(await res.json(), []);

    res = await fetch(`${base}/api/builds`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ name: "checkout-service", branch: "main" }),
    });
    assert.equal(res.status, 201);
    const created = await res.json();
    assert.equal(created.name, "checkout-service");
    assert.equal(created.status, "queued");
    assert.equal(created.id, 1);

    res = await fetch(`${base}/api/builds`);
    const list = await res.json();
    assert.equal(list.length, 1);
  } finally {
    server.close();
  }
});

test("creating a build without a name fails", async () => {
  const { server, base } = startServer();
  try {
    const res = await fetch(`${base}/api/builds`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ branch: "main" }),
    });
    assert.equal(res.status, 400);
  } finally {
    server.close();
  }
});

test("advancing a build moves it through the status flow", async () => {
  const { server, base } = startServer();
  try {
    let res = await fetch(`${base}/api/builds`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ name: "worker", branch: "main" }),
    });
    const build = await res.json();

    res = await fetch(`${base}/api/builds/${build.id}/advance`, { method: "POST" });
    assert.equal((await res.json()).status, "running");

    res = await fetch(`${base}/api/builds/${build.id}/advance`, { method: "POST" });
    assert.equal((await res.json()).status, "succeeded");

    res = await fetch(`${base}/api/builds/999/advance`, { method: "POST" });
    assert.equal(res.status, 404);
  } finally {
    server.close();
  }
});

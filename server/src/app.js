import express from "express";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { createStore } from "./store.js";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/**
 * Build the Express application. Accepts a store so tests can inject a fresh
 * one; defaults to a seeded in-memory store for normal use.
 */
export function createApp({ store = createStore(), serveClient = true } = {}) {
  const app = express();
  app.use(express.json());

  app.get("/api/health", (_req, res) => {
    res.json({ status: "ok", uptime: process.uptime() });
  });

  app.get("/api/builds", (_req, res) => {
    res.json(store.list());
  });

  app.post("/api/builds", (req, res) => {
    const { name, branch } = req.body ?? {};
    if (!name || typeof name !== "string" || !name.trim()) {
      return res.status(400).json({ error: "name is required" });
    }
    const build = store.create({ name, branch });
    res.status(201).json(build);
  });

  app.post("/api/builds/:id/advance", (req, res) => {
    const build = store.advance(req.params.id);
    if (!build) return res.status(404).json({ error: "build not found" });
    res.json(build);
  });

  if (serveClient) {
    const clientDir = path.resolve(__dirname, "../../web/dist");
    app.use(express.static(clientDir));
    app.get(/^(?!\/api\/).*/, (_req, res) => {
      res.sendFile(path.join(clientDir, "index.html"));
    });
  }

  return app;
}

import { createApp } from "./app.js";
import { createStore } from "./store.js";

const PORT = Number(process.env.PORT) || 3001;

const store = createStore();
store.seed();

const app = createApp({ store });

app.listen(PORT, () => {
  console.log(`[agentic-build] API listening on http://localhost:${PORT}`);
});

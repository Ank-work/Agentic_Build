import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { test } from "node:test";
import { fileURLToPath } from "node:url";
import { applyDefaults } from "./apply-defaults.mjs";

const here = dirname(fileURLToPath(import.meta.url));

function loadJSON(rel) {
  return JSON.parse(readFileSync(join(here, rel), "utf8"));
}

function validate(schema, data, path = "$") {
  if (schema.const !== undefined && data !== schema.const) {
    throw new Error(`${path}: expected const ${JSON.stringify(schema.const)}`);
  }
  if (schema.enum && !schema.enum.includes(data)) {
    throw new Error(`${path}: value not in enum`);
  }
  if (schema.type === "object") {
    if (data === null || typeof data !== "object" || Array.isArray(data)) {
      throw new Error(`${path}: expected object`);
    }
    for (const key of schema.required || []) {
      if (!Object.prototype.hasOwnProperty.call(data, key)) {
        throw new Error(`${path}: missing required ${key}`);
      }
    }
    const props = schema.properties || {};
    if (schema.additionalProperties === false) {
      for (const key of Object.keys(data)) {
        if (!Object.prototype.hasOwnProperty.call(props, key)) {
          throw new Error(`${path}: unknown property ${key}`);
        }
      }
    } else if (schema.additionalProperties && typeof schema.additionalProperties === "object") {
      for (const [key, val] of Object.entries(data)) {
        if (!Object.prototype.hasOwnProperty.call(props, key)) {
          validate(schema.additionalProperties, val, `${path}.${key}`);
        }
      }
    }
    for (const [key, sub] of Object.entries(props)) {
      if (Object.prototype.hasOwnProperty.call(data, key)) {
        validate(sub, data[key], `${path}.${key}`);
      }
    }
    return;
  }
  if (schema.type === "array") {
    if (!Array.isArray(data)) {
      throw new Error(`${path}: expected array, never a shell string`);
    }
    if (schema.minItems != null && data.length < schema.minItems) {
      throw new Error(`${path}: minItems`);
    }
    if (schema.items) {
      data.forEach((item, i) => validate(schema.items, item, `${path}[${i}]`));
    }
    return;
  }
  if (schema.type === "string") {
    if (typeof data !== "string") {
      throw new Error(`${path}: expected string`);
    }
    if (schema.minLength != null && data.length < schema.minLength) {
      throw new Error(`${path}: minLength`);
    }
    return;
  }
  if (schema.type === "integer") {
    if (typeof data !== "number" || !Number.isInteger(data)) {
      throw new Error(`${path}: expected integer`);
    }
    if (schema.minimum != null && data < schema.minimum) {
      throw new Error(`${path}: minimum`);
    }
    if (schema.maximum != null && data > schema.maximum) {
      throw new Error(`${path}: maximum`);
    }
  }
}

const schema = loadJSON("job-spec.v1.json");

test("schema parses and $id/title/required match §6.1", () => {
  assert.equal(schema.$id, "https://aeon.local/schemas/job-spec.v1.json");
  assert.equal(schema.title, "AeonJobSpec");
  assert.deepEqual(schema.required, ["apiVersion", "kind", "metadata", "spec"]);
  assert.equal(schema.additionalProperties, false);
  assert.equal(schema.properties.apiVersion.const, "aeon.dev/v1");
  assert.equal(schema.properties.kind.const, "Job");
});

test("argv as a shell string is invalid", () => {
  const job = loadJSON("fixtures/invalid-shell-string.json");
  assert.equal(typeof job.spec.argv, "string");
  assert.throws(() => validate(schema, job), /expected array|shell string/i);
});

test("extra spec.command / spec.shell is invalid", () => {
  const job = loadJSON("fixtures/invalid-shell-command-field.json");
  assert.throws(() => validate(schema, job), /unknown property (command|shell)/);
});

test("applyDefaults sets missing network / network.mode to none", () => {
  const job = loadJSON("fixtures/missing-network.json");
  assert.equal(job.spec.network, undefined);
  const withDefaults = applyDefaults(job);
  assert.equal(withDefaults.spec.network.mode, "none");
  validate(schema, withDefaults);
});

test("valid fixture has argv as a non-empty string array", () => {
  const job = loadJSON("fixtures/valid-minimal.json");
  assert.ok(Array.isArray(job.spec.argv));
  assert.ok(job.spec.argv.length >= 1);
  assert.ok(job.spec.argv.every((a) => typeof a === "string"));
  validate(schema, job);
});

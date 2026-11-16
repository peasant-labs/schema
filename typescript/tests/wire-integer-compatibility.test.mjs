import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readFile } from "node:fs/promises";
import test from "node:test";
import { parse } from "yaml";

import {
  TypesVersion,
  parseSessionDetailPayloadText,
  parseTranscriptContentText,
  zAuthoritativePublishRequest,
  zPublishRequest,
  zRetainedUnknownRecord,
  zSessionDetailPayload,
  zTimestampInfo,
  zTranscriptSessionHeader,
  zVillageAvailableRepository,
} from "../dist/index.js";
import { Classification, loadCorpus } from "../dist/testcase.js";

const base = "../../testdata/typescript/wire_integer_compatibility";
const raw = await readFile(new URL(`${base}.yaml`, import.meta.url), "utf8");
const manifest = parse(await readFile(new URL(`${base}_manifest.yaml`, import.meta.url), "utf8"), { uniqueKeys: true });
const cases = loadCorpus(raw, { decodeInput: (value) => value, decodeExpected: (value) => value }).cases;
assert.deepEqual(Object.keys(manifest), ["requiredNames"]);
assert.ok(manifest.requiredNames.length > 0);
assert.equal(new Set(manifest.requiredNames).size, manifest.requiredNames.length);
assert.deepEqual(cases.map((row) => row.name).sort(), [...manifest.requiredNames].sort());

const publishCorpus = parse(await readFile(new URL("../../testdata/publish/verdicts.yaml", import.meta.url), "utf8"));
const fullPublish = publishCorpus.cases.find((row) => row.name === "full-valid");
assert.ok(fullPublish, "the existing full publish fixture must be present");
const authoritativeCorpus = parse(await readFile(new URL("../../testdata/publication/contract.yaml", import.meta.url), "utf8"));
const fullAuthoritative = authoritativeCorpus.parent_identity.cases.find((row) => row.name === "root-omits-parent");
assert.ok(fullAuthoritative, "the existing authoritative publish fixture must be present");

const schemas = Object.freeze({
  timestamp: zTimestampInfo,
  publish: zPublishRequest,
  authoritative: zAuthoritativePublishRequest,
  repository: zVillageAvailableRepository,
  retained: zRetainedUnknownRecord,
  header: zTranscriptSessionHeader,
  durable: zSessionDetailPayload,
});

test("generated integer validators retain formatted compatibility and served-header parity", async (t) => {
  for (const row of cases) {
    await t.test(row.name, () => {
      const schema = schemas[row.input.target];
      assert.ok(schema, `unknown integer compatibility target ${row.input.target}`);
      const value = JSON.parse(row.input.json);
      if (row.input.bigint) {
        value.start = BigInt(value.start);
        value.end = BigInt(value.end);
      }
      let input = value;
      if (row.input.path !== undefined) {
        input = row.input.target === "authoritative" ? JSON.parse(fullAuthoritative.input) :
          row.input.target === "publish" ? JSON.parse(fullPublish.body) : JSON.parse(fullPublish.body).timestamp;
        const parts = row.input.path.split(".");
        let cursor = input;
        for (const part of parts.slice(0, -1)) cursor = cursor[part];
        cursor[parts.at(-1)] = value;
      } else if (row.input.target === "publish") {
        input = { ...JSON.parse(fullPublish.body), timestamp: value };
      }
      const result = schema.safeParse(input);
      assert.equal(row.classification, row.expected.valid ? Classification.MustPass : Classification.MustFail);
      assert.equal(result.success, row.expected.valid, result.success ? "" : result.error.issues.map((issue) => issue.message).join("; "));
      if (row.input.target === "durable") {
        const envelope = `{"contractVersion":"0.1.1","kind":"session_detail","sessionDetail":${row.input.json}}`;
        if (row.expected.valid) {
          assert.deepEqual(parseSessionDetailPayloadText(row.input.json), result.data);
          assert.deepEqual(parseTranscriptContentText(envelope).sessionDetail, result.data);
        } else {
          assert.throws(() => parseSessionDetailPayloadText(row.input.json));
          assert.throws(() => parseTranscriptContentText(envelope));
        }
      }
      if (result.success && row.expected.start !== undefined) {
        const timestamp = row.input.target === "publish" ? result.data.timestamp : result.data;
        assert.equal(typeof timestamp.start, "bigint");
        assert.equal(typeof timestamp.end, "bigint");
        assert.equal(timestamp.start.toString(), row.expected.start);
        assert.equal(timestamp.end.toString(), row.expected.end);
      }
      if (result.success && row.expected.numberPosition) assert.equal(typeof result.data.position, "number");
      if (result.success && row.input.target === "header") assert.deepEqual(result.data, input);
    });
  }
});

test("the adjacent endpoint lexemes are indistinguishable after JSON.parse", () => {
  const values = new Map(cases.filter((row) => row.input.target === "header")
    .map((row) => [row.name, JSON.parse(row.input.json).conversationTurnCount]));
  assert.equal(values.get("header-upper-endpoint-accepted"), values.get("header-upper-ambiguous-overflow-indistinguishable"));
  assert.equal(values.get("header-lower-endpoint-accepted"), values.get("header-lower-ambiguous-overflow-indistinguishable"));
});

test("existing formatted and publication validators match the pre-change source", async () => {
  const baseline = parse(await readFile(new URL("../../testdata/typescript/wire_validator_baseline.yaml", import.meta.url), "utf8"), { uniqueKeys: true });
  const source = await readFile(new URL("../src/internal/generated/contract/zod.gen.ts", import.meta.url), "utf8");
  const spec = JSON.parse(await readFile(new URL(`../../generated/types-${TypesVersion}.json`, import.meta.url), "utf8")).components.schemas;
  assert.match(baseline.capturedFrom, /^[0-9a-f]{40}$/);
  for (const section of [baseline.publication, baseline.formatted]) {
    assert.ok(section.requiredNames.length > 0);
    assert.equal(new Set(section.requiredNames).size, section.requiredNames.length);
  }
  const declaration = (name) => {
    const prefix = `export const z${name} = `;
    assert.equal(source.split(prefix).length, 2, `one generated declaration is required for ${name}`);
    const [body, rest] = source.split(prefix)[1].split(`\n\nexport type ${name} =`);
    assert.notEqual(rest, undefined, `the generated type is required for ${name}`);
    return body;
  };
  const hash = (text) => createHash("sha256").update(text).digest("hex");

  const reached = new Set();
  const pending = ["PublishRequest", "AuthoritativePublishRequest"];
  while (pending.length > 0) {
    const name = pending.pop();
    if (reached.has(name)) continue;
    assert.ok(spec[name], `missing publication component ${name}`);
    reached.add(name);
    const visit = (node) => {
      if (typeof node !== "object" || node === null) return;
      if (typeof node.$ref === "string") {
        assert.ok(node.$ref.startsWith("#/components/schemas/"), node.$ref);
        pending.push(node.$ref.slice("#/components/schemas/".length));
      }
      for (const child of Object.values(node)) visit(child);
    };
    visit(spec[name]);
  }
  assert.deepEqual([...reached].sort(), [...baseline.publication.requiredNames].sort());
  const publication = baseline.publication.requiredNames.map((name) => `${name}\n${declaration(name)}\n`).join("");
  assert.equal(hash(publication), baseline.publication.sha256);
  assert.notEqual(hash(`${publication}mutation`), baseline.publication.sha256);

  const formatted = baseline.formatted.requiredNames.map((qualified) => {
    const [name, field] = qualified.split(".");
    const property = spec[name]?.properties?.[field];
    assert.ok(property, `missing formatted property ${qualified}`);
    assert.ok([property, ...(property.anyOf ?? []), ...(property.oneOf ?? [])]
      .some((variant) => variant.type === "integer" && ["int32", "int64"].includes(variant.format)), qualified);
    const escaped = field.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    const line = declaration(name).match(new RegExp(`^\\s*${escaped}: [^\\n]+`, "m"))?.[0];
    assert.ok(line, `missing formatted validator ${qualified}`);
    return `${qualified}\n${line.trim()}\n`;
  }).join("");
  assert.equal(hash(formatted), baseline.formatted.sha256);
});

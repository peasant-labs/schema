import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { parse } from "yaml";

import * as root from "../dist/index.js";
import * as localApi from "../dist/local-api.js";
import * as villageApi from "../dist/village-api.js";
import { loadCorpus } from "../dist/testcase.js";

// The route members are a projection of the built Village API and Local API
// documents. These tests read the documents themselves, so a member without an
// operation, or an operation without a member, fails here as well as in the
// type test (route-members.ts) and the Go test (route/route_test.go).

const apis = [
  {
    label: "village",
    module: villageApi,
    members: villageApi.VillageRoute,
    all: villageApi.ALL_VILLAGE_ROUTES,
    document: JSON.parse(await readFile(new URL(`../../generated/village-api-${villageApi.VillageAPIVersion}.json`, import.meta.url), "utf8")),
  },
  {
    label: "local",
    module: localApi,
    members: localApi.LocalRoute,
    all: localApi.ALL_LOCAL_ROUTES,
    document: JSON.parse(await readFile(new URL(`../../generated/peasantlocal-api-${localApi.PeasantLocalAPIVersion}.json`, import.meta.url), "utf8")),
  },
];

function documentOperations(document) {
  const keys = [];
  for (const [pattern, item] of Object.entries(document.paths)) {
    for (const [method, operation] of Object.entries(item)) {
      if (method === "parameters") continue;
      keys.push(`${method.toUpperCase()} ${pattern} ${operation.operationId}`);
    }
  }
  return keys.sort();
}

const memberKey = (member) => `${member.method} ${member.pattern} ${member.operationId}`;

test("each API's members are exactly the operations of its document", async (t) => {
  for (const api of apis) {
    await t.test(api.label, () => {
      const members = Object.values(api.members);
      assert.deepEqual(members.map(memberKey).sort(), documentOperations(api.document));
      assert.equal(new Set(members.map(memberKey)).size, members.length, "two members share one operation");
      for (const [name, member] of Object.entries(api.members)) {
        assert.equal(name, member.operationId[0].toUpperCase() + member.operationId.slice(1), `member ${name} is not named by its operationId`);
        assert.ok(villageApi.isHttpMethod(member.method), `member ${name} has method ${member.method} outside ALL_HTTP_METHODS`);
      }
    });
  }
});

test("each ALL_*_ROUTES set holds every member once, by pattern and then by method", async (t) => {
  const rank = new Map(villageApi.ALL_HTTP_METHODS.map((method, index) => [method, index]));
  for (const api of apis) {
    await t.test(api.label, () => {
      const members = Object.values(api.members);
      assert.equal(api.all.length, members.length);
      assert.deepEqual(new Set(api.all), new Set(members), "ALL_*_ROUTES must hold the member objects themselves");
      const sorted = [...api.all].sort((a, b) => (a.pattern === b.pattern ? rank.get(a.method) - rank.get(b.method) : a.pattern < b.pattern ? -1 : 1));
      assert.deepEqual(api.all, sorted);
    });
  }
});

test("members, their sets and the method set are frozen", () => {
  for (const value of [villageApi.HttpMethod, villageApi.ALL_HTTP_METHODS, ...apis.flatMap((api) => [api.members, api.all, ...api.all])]) {
    assert.ok(Object.isFrozen(value));
  }
  assert.throws(() => {
    villageApi.VillageRoute.GetSchemaVersion.pattern = "/api/v1/other";
  }, TypeError);
});

test("both subpaths share one method set, and the root does not carry route members", () => {
  assert.equal(localApi.HttpMethod, villageApi.HttpMethod);
  assert.equal(localApi.ALL_HTTP_METHODS, villageApi.ALL_HTTP_METHODS);
  assert.equal(localApi.isHttpMethod, villageApi.isHttpMethod);
  assert.deepEqual([...villageApi.ALL_HTTP_METHODS], Object.values(villageApi.HttpMethod));
  assert.equal(villageApi.isHttpMethod("HEAD"), false);
  assert.equal(villageApi.isHttpMethod("get"), false);
  for (const name of ["VillageRoute", "LocalRoute", "ALL_VILLAGE_ROUTES", "ALL_LOCAL_ROUTES", "HttpMethod", "ALL_HTTP_METHODS", "isHttpMethod"]) {
    assert.equal(Object.hasOwn(root, name), false, `${name} unexpectedly exported from the root facade`);
  }
  assert.equal(Object.hasOwn(villageApi, "LocalRoute"), false);
  assert.equal(Object.hasOwn(localApi, "VillageRoute"), false);
});

function member(apiLabel, name, path) {
  const api = apis.find((candidate) => candidate.label === apiLabel);
  if (api === undefined) throw new TypeError(`${path}.api: ${JSON.stringify(apiLabel)} is not village or local`);
  const found = api.members[name];
  if (found === undefined) throw new TypeError(`${path}.member: ${JSON.stringify(name)} is not a ${apiLabel} route member`);
  return found;
}

function decodePathInput(value, path) {
  if (typeof value !== "object" || value === null || Array.isArray(value)) throw new TypeError(`${path}: must be a mapping`);
  for (const key of Object.keys(value)) {
    if (!["api", "member", "params"].includes(key)) throw new TypeError(`${path}.${key}: unknown key`);
  }
  if (value.params !== undefined && value.params !== null && (typeof value.params !== "object" || Array.isArray(value.params))) throw new TypeError(`${path}.params: must be a mapping or null when present`);
  return { member: member(value.api, value.member, path), params: value.params };
}

async function loadRouteCorpus(fixture, manifest, decodeExpected) {
  const corpus = loadCorpus(await readFile(new URL(fixture, import.meta.url), "utf8"), { decodeInput: decodePathInput, decodeExpected });
  const required = parse(await readFile(new URL(manifest, import.meta.url), "utf8")).requiredNames;
  assert.deepEqual(corpus.cases.map((testCase) => testCase.name).sort(), [...required].sort(), `${fixture}: a case was added or removed without its manifest entry`);
  return corpus;
}

test("path() builds the shared fixture paths", async (t) => {
  const corpus = await loadRouteCorpus("../../testdata/route_paths.yaml", "../../testdata/route_paths_manifest.yaml", (value, path) => {
    if (typeof value?.typescript !== "string" || typeof value?.go !== "string") throw new TypeError(`${path}: must give the go and typescript paths`);
    return value.typescript;
  });
  for (const testCase of corpus.cases) {
    await t.test(testCase.name, () => {
      assert.equal(testCase.classification, "must-pass");
      const { member: route, params } = testCase.input;
      assert.equal(params === undefined ? route.path() : route.path(params), testCase.expected);
    });
  }
});

test("path() refuses a value a JavaScript caller passes outside the type", async (t) => {
  const corpus = await loadRouteCorpus("../../testdata/typescript/route_path_refusals.yaml", "../../testdata/typescript/route_path_refusals_manifest.yaml", (value, path) => {
    if (typeof value?.error_contains !== "string") throw new TypeError(`${path}.error_contains: must be a string`);
    return value.error_contains;
  });
  for (const testCase of corpus.cases) {
    await t.test(testCase.name, () => {
      assert.equal(testCase.classification, "must-fail");
      const { member: route, params } = testCase.input;
      assert.throws(() => (params === undefined ? route.path() : route.path(params)), (error) => {
        assert.ok(error instanceof TypeError);
        assert.ok(error.message.includes(testCase.expected), `${error.message} does not contain ${testCase.expected}`);
        assert.ok(error.message.includes(route.operationId) && error.message.includes("pass "), `${error.message} does not name the operation and the fix`);
        return true;
      });
    });
  }
});

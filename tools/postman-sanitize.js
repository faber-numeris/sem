#!/usr/bin/env node
// Prepares a Postman collection for postman-to-openapi (p2o).
//
// It fixes collection defects that either break p2o or produce a spec that
// OpenAPI tooling (e.g. oapi-codegen) rejects:
//
//   - response examples declaring a JSON content type with a body that is not
//     valid JSON (placeholders such as `"interval": integer`) are re-labelled
//     as text/plain, so they are kept verbatim instead of aborting p2o;
//   - response examples without a status code (p2o emits an "undefined"
//     response key) get one, taken from names like "Status201-..." or 200;
//   - parameter/header keys with surrounding whitespace are trimmed;
//   - trailing slashes are dropped from request paths, so "/query" and
//     "/query/" don't become two operations with the same generated name.
//
// Usage: node tools/postman-sanitize.js <collection.json> <output.json>

const fs = require("fs");

const [source, output] = process.argv.slice(2);
if (!source || !output) {
  console.error("usage: postman-sanitize.js <collection.json> <output.json>");
  process.exit(1);
}

const collection = JSON.parse(fs.readFileSync(source, "utf8"));
const fixed = [];
const coded = [];
const trimmed = [];
const slashed = [];

const isJsonContentType = (value) => /json/i.test(value || "");

function ensureStatusCode(itemName, response) {
  if (Number.isInteger(response.code)) return;
  const match = /\b([1-5]\d\d)\b/.exec(response.name || "");
  response.code = match ? Number(match[1]) : 200;
  response.status = response.status || (response.code === 200 ? "OK" : "");
  coded.push(`${itemName} > ${response.name} (${response.code})`);
}

function trimKeys(itemName, list) {
  for (const entry of Array.isArray(list) ? list : []) {
    if (entry && typeof entry.key === "string" && entry.key !== entry.key.trim()) {
      trimmed.push(`${itemName} > ${JSON.stringify(entry.key)}`);
      entry.key = entry.key.trim();
    }
  }
}

function sanitizeRequest(item) {
  const request = item.request;
  if (!request || typeof request !== "object") return;
  trimKeys(item.name, request.header);
  if (request.url && typeof request.url === "object") {
    const path = request.url.path;
    if (Array.isArray(path) && path.length > 1 && path[path.length - 1] === "") {
      while (path.length > 1 && path[path.length - 1] === "") path.pop();
      if (typeof request.url.raw === "string") {
        request.url.raw = request.url.raw.replace(/\/+(?=($|\?))/, "");
      }
      slashed.push(item.name);
    }
    trimKeys(item.name, request.url.query);
    trimKeys(item.name, request.url.variable);
  }
  if (request.body) {
    trimKeys(item.name, request.body.formdata);
    trimKeys(item.name, request.body.urlencoded);
  }
}

function sanitizeResponse(response) {
  const headers = Array.isArray(response.header) ? response.header : [];
  const contentType = headers.find(
    (h) => h && typeof h.key === "string" && h.key.toLowerCase() === "content-type",
  );
  const declaredJson =
    isJsonContentType(contentType && contentType.value) ||
    response._postman_previewlanguage === "json";
  const body = response.body || "";
  if (!declaredJson || body.trim() === "") return;

  try {
    JSON.parse(body);
  } catch {
    if (contentType) {
      contentType.value = "text/plain";
    } else {
      headers.push({ key: "Content-Type", value: "text/plain" });
      response.header = headers;
    }
    response._postman_previewlanguage = "text";
    fixed.push(response.name);
  }
}

function walk(items) {
  for (const item of items || []) {
    walk(item.item);
    sanitizeRequest(item);
    for (const response of item.response || []) {
      ensureStatusCode(item.name, response);
      sanitizeResponse(response);
    }
  }
}

walk(collection.item);
fs.writeFileSync(output, JSON.stringify(collection));

function report(list, message) {
  if (list.length === 0) return;
  console.error(`${message} (${list.length}):`);
  for (const name of list) console.error(`  - ${name}`);
}

report(fixed, "Re-labelled invalid JSON examples as text/plain");
report(coded, "Assigned status codes to examples without one");
report(trimmed, "Trimmed whitespace from parameter keys");
report(slashed, "Dropped trailing slashes from request paths");

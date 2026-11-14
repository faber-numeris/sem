#!/usr/bin/env node
// Converts a Postman collection into a Bruno OpenCollection (YAML) folder.
//
// The published `bru import` command does not support Postman yet, so this
// script combines the building blocks it ships with: `postmanToBruno` from
// @usebruno/converters and the collection writer from @usebruno/cli.
//
// Usage (see the `postman-to-bruno` recipe in the justfile):
//   node tools/postman-to-bruno.js <postman.json> <output-dir> [collection-name]
//
// Requires @usebruno/cli (and its @usebruno/converters dependency) to be
// resolvable, e.g. via NODE_PATH.

const fs = require('fs');
const path = require('path');
const { postmanToBruno } = require('@usebruno/converters');
const { createCollectionFromBrunoObject } = require('@usebruno/cli/src/utils/collection');

const main = async () => {
  const [source, output, collectionName] = process.argv.slice(2);
  if (!source || !output) {
    console.error('Usage: postman-to-bruno.js <postman.json> <output-dir> [collection-name]');
    process.exit(2);
  }

  const postman = JSON.parse(fs.readFileSync(source, 'utf8').replace(/^﻿/, ''));
  const result = await postmanToBruno(postman);
  // Newer converter versions return { collection, issues }; older ones return the collection.
  const collection = result.collection || result;
  if (collectionName) {
    collection.name = collectionName;
  }

  const outputDir = path.resolve(output);
  if (fs.existsSync(outputDir) && fs.readdirSync(outputDir).length > 0) {
    console.error(`Output directory is not empty: ${outputDir}`);
    process.exit(1);
  }
  fs.mkdirSync(outputDir, { recursive: true });

  await createCollectionFromBrunoObject(collection, outputDir, { format: 'yml' });
  const patched = translateLeftoverPmCalls(outputDir);
  console.log(`Bruno OpenCollection created at ${outputDir} (${patched} files had leftover pm.* calls translated)`);

  const leftovers = findLeftoverPmCalls(outputDir);
  if (leftovers.length > 0) {
    console.warn('Warning: untranslated pm.* calls remain in:');
    leftovers.forEach((f) => console.warn(`  ${path.relative(outputDir, f)}`));
  }
};

// The converter leaves a few Postman APIs untouched; map the ones this
// collection uses onto their Bruno equivalents.
const PM_REPLACEMENTS = [
  [
    /bru\.getEnvName\(\) \? pm\.environment : pm\.collectionVariables/g,
    'bru.getEnvName() ? { set: (k, v) => bru.setEnvVar(k, v), get: (k) => bru.getEnvVar(k) } : { set: (k, v) => bru.setVar(k, v), get: (k) => bru.getVar(k) }'
  ],
  [/\bpm\.variables\.set\(/g, 'bru.setVar('],
  [/\bpm\.variables\.get\(/g, 'bru.getVar('],
  [/\bpm\.environment\.set\(/g, 'bru.setEnvVar('],
  [/\bpm\.environment\.get\(/g, 'bru.getEnvVar('],
  [/\bpm\.collectionVariables\.set\(/g, 'bru.setVar('],
  [/\bpm\.collectionVariables\.get\(/g, 'bru.getVar(']
];

const listYamlFiles = (dir) =>
  fs.readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) return listYamlFiles(full);
    return entry.name.endsWith('.yml') ? [full] : [];
  });

const translateLeftoverPmCalls = (dir) => {
  let count = 0;
  for (const file of listYamlFiles(dir)) {
    const original = fs.readFileSync(file, 'utf8');
    const updated = PM_REPLACEMENTS.reduce((text, [pattern, replacement]) => text.replace(pattern, replacement), original);
    if (updated !== original) {
      fs.writeFileSync(file, updated);
      count++;
    }
  }
  return count;
};

const findLeftoverPmCalls = (dir) =>
  listYamlFiles(dir).filter((file) => /\bpm\./.test(fs.readFileSync(file, 'utf8')));

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

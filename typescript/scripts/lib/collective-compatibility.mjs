// Go's explicit retained aliases own the rename window. Derive the npm aliases
// from those actual declarations, not from a second list of wire definitions.
export function collectiveCompatibilityAliases(source, catalog) {
  const inventories = new Map([...source.matchAll(/^var (AllVillage\w+) = (AllVillage\w+)$/gm)].map(match => [match[1], match[2]]));
  const enums = new Map(catalog.enums.map(row => [row.name, row]));
  const aliases = [...source.matchAll(/^type (Village\w+) = (Village\w+)$/gm)].map(match => {
    const row = { name: match[1], canonical: match[2] }, enumRow = enums.get(match[2]);
    if (enumRow !== undefined) {
      row.all_name = inventories.get(enumRow.all_name);
      row.canonical_all_name = enumRow.all_name;
      if (row.all_name === undefined) throw new Error(`Collective compatibility generation cannot find the retained Go inventory for ${row.name}; npm compatibility would be incomplete; restore the Go alias or update the source parser before generating.`);
    }
    return row;
  });
  if (aliases.length === 0 || new Set(aliases.map(row => row.name)).size !== aliases.length) throw new Error("Collective compatibility generation requires unique retained Go aliases; inspect collective_compatibility.go before publishing.");
  return aliases.sort((a, b) => a.name.localeCompare(b.name, "en"));
}

export function renderCollectiveCompatibility(source, catalog) {
  const declarations = collectiveCompatibilityAliases(source, catalog).flatMap(row => {
    const deprecated = target => `/** @deprecated Use ${target}; retained for the complete alias release window. */\n`;
    const lines = [deprecated(`z${row.canonical}`) + `export { z${row.canonical} as z${row.name} } from "./contract/zod.gen.js";`];
    if (row.all_name === undefined) lines.push(deprecated(row.canonical) + `export type { ${row.canonical} as ${row.name} } from "./contract/zod.gen.js";`);
    else {
      lines.push(deprecated(row.canonical) + `export { ${row.canonical} as ${row.name} } from "./enums.gen.js";`);
      lines.push(deprecated(row.canonical_all_name) + `export { ${row.canonical_all_name} as ${row.all_name} } from "./enums.gen.js";`);
      lines.push(deprecated(`is${row.canonical}`) + `export { is${row.canonical} as is${row.name} } from "./enums.gen.js";`);
    }
    return lines;
  });
  return "// Code generated from retained Go collective aliases. DO NOT EDIT.\n" + declarations.join("\n") + "\n";
}

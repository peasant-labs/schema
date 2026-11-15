// The generator's integer and date-time defaults are narrower than served
// values. Preserve declared safe bounds and constrain wider integer numbers to
// the represented int64 domain; fail if emitted property markers change.
// Without a declared safe maximum, unformatted integers use
// z.number().refine(Number.isInteger) and inclusive represented bounds
// -9223372036854775808..9223372036854775808.
// JSON.parse rounds numeric lexemes: adjacent out-of-range lexemes can share
// an endpoint. The Go constructor and served bytes retain the exact bounds;
// consumers treat unsafe numbers as unknown, not counts, order, or totals.
// Evidence timestamps have their own UTC refinement.
// Existing publication schemas and their referenced components retain their
// prior validators; the caller identifies those roots from the canonical spec.
export function applyWireScalarZodRefinements(source, spec, { preserveComponentsReachableFrom = [] } = {}) {
  let output = source;
  const preserved = referencedComponents(spec.components.schemas, preserveComponentsReachableFrom);
  for (const [name, component] of Object.entries(spec.components.schemas)) {
    if (preserved.has(name)) continue;
    const fields = new Map();
    const collect = (node) => {
      if (typeof node !== "object" || node === null) return;
      for (const [field, property] of Object.entries(node.properties ?? {})) {
        const scalar = scalarRule(property);
        if (scalar !== undefined) {
          const list = fields.get(field) ?? [];
          list.push(scalar);
          fields.set(field, list);
        }
        collect(property);
      }
      for (const variant of [...(node.oneOf ?? []), ...(node.anyOf ?? []), ...(node.allOf ?? [])]) collect(variant);
    };
    collect(component);
    if (fields.size === 0) continue;
    const start = output.indexOf(`export const z${name} = `);
    const end = output.indexOf(`\n\nexport type ${name} =`, start);
    if (start < 0 || end < start) {
      throw new Error(`Wire scalar generation lost ${name}; inspect the pinned generator before publishing.`);
    }
    let declaration = output.slice(start, end);
    for (const [field, rules] of fields) {
      const escaped = field.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
      const marker = new RegExp(`^(\\s*(?:${escaped}|"${escaped}"|'${escaped}'): )([^\\n]+)$`, "gm");
      let index = 0;
      declaration = declaration.replace(marker, (line, prefix, expression) => {
        const rule = rules[index++];
        if (rule === undefined) throw new Error(`Wire scalar generation found extra ${name}.${field} markers; inspect the pinned generator before publishing.`);
        const context = `${name}.${field}`;
        if (rule.kind === "date") {
          if (!expression.includes("z.iso.datetime()")) throw new Error(`Wire scalar generation lost date-time marker ${context}; inspect the pinned generator before publishing.`);
          return prefix + expression.replace("z.iso.datetime()", "z.iso.datetime({ offset: true })");
        }
        return prefix + integerExpression(expression, rule.schema, context);
      });
      if (index !== rules.length) throw new Error(`Wire scalar generation lost ${name}.${field} markers (${index}/${rules.length}); inspect the pinned generator before publishing.`);
    }
    output = output.slice(0, start) + declaration + output.slice(end);
  }
  return output;
}

function referencedComponents(schemas, roots) {
  const preserved = new Set();
  const visitComponent = (name) => {
    if (!Object.hasOwn(schemas, name)) throw new Error(`Wire scalar generation found missing protected component ${name}; inspect the canonical publication schemas before publishing.`);
    if (preserved.has(name)) return;
    preserved.add(name);
    visitNode(schemas[name]);
  };
  const visitNode = (node) => {
    if (typeof node !== "object" || node === null) return;
    if (typeof node.$ref === "string") {
      const prefix = "#/components/schemas/";
      if (!node.$ref.startsWith(prefix) || node.$ref.slice(prefix.length).includes("/")) {
        throw new Error(`Wire scalar generation found unsupported protected schema reference ${node.$ref}; inspect the canonical publication schemas before publishing.`);
      }
      visitComponent(node.$ref.slice(prefix.length));
    }
    for (const child of Object.values(node)) visitNode(child);
  };
  for (const name of roots) visitComponent(name);
  return preserved;
}

function scalarRule(property) {
  const variants = [property, ...(property.anyOf ?? []), ...(property.oneOf ?? [])];
  const matches = variants.filter((item) => {
    const types = Array.isArray(item.type) ? item.type : [item.type];
    return (types.includes("integer") && item.format === undefined) ||
      (item.format === "date-time" && item.not?.pattern === undefined);
  });
  if (matches.length > 1) throw new Error("Wire scalar generation found multiple scalar variants on one property; inspect the schema before publishing.");
  if (matches.length === 0) return undefined;
  const schema = matches[0];
  return { kind: schema.format === "date-time" ? "date" : "integer", schema };
}

function integerExpression(expression, schema, context) {
  const max = schema.maximum;
  const safeBound = max !== undefined && max <= Number.MAX_SAFE_INTEGER;
  if (expression.startsWith("z.int()")) {
    if (schema.minimum !== undefined && !expression.includes(`.gte(${schema.minimum})`)) {
      throw new Error(`Wire scalar generation lost minimum marker ${context}; inspect the pinned generator before publishing.`);
    }
    if (max !== undefined && !expression.includes(`.lte(${max})`)) {
      throw new Error(`Wire scalar generation lost maximum marker ${context}; inspect the pinned generator before publishing.`);
    }
    if (safeBound) return expression;

    // JSON.parse rounds the positive int64 maximum to 2^63. Inclusive bounds
    // admit that served endpoint, although they cannot distinguish an adjacent
    // overflow lexeme that rounds to the same number. Go checks exact lexemes.
    const lower = -9223372036854775808;
    const upper = 9223372036854775808;
    const suffix = expression.match(/((?:\.(?:optional|nullish|nullable)\(\))*)(,?)$/);
    if (suffix === null) throw new Error(`Wire scalar generation lost integer suffix ${context}; inspect the pinned generator before publishing.`);
    let result = expression.slice(0, expression.length - suffix[0].length).replace("z.int()", "z.number().refine(Number.isInteger)");
    if (schema.minimum === undefined || schema.minimum < lower) result += ".gte(-9223372036854775808)";
    if (max === undefined || max > upper) result += ".lte(9223372036854775808)";
    return result + suffix[0];
  }
  throw new Error(`Wire scalar generation lost integer marker ${context}; inspect the pinned generator before publishing.`);
}

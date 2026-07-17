import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const packageRoot = dirname(fileURLToPath(import.meta.url));
const moduleRoot = join(packageRoot, "..");
const versionsSource = readFileSync(join(moduleRoot, "versions.go"), "utf8");
const typesVersion = requiredMatch(versionsSource, /TypesVersion\s*=\s*"([^"]+)"/, "TypesVersion", "versions.go");

// canonicalTypeName is the identity function: Hey API's `case: "preserve"`
// (set below on both `definitions.name` and `types.infer.name`) already emits
// each Types OpenAPI component's raw name unchanged, and the Go OpenAPI
// builder is the one place that decides those names (verified: it names
// acronym-bearing components CLILoginQuery/ModelID/SessionID/TranscriptID
// correctly cased already). A hand-maintained rename lookup table here would
// be redundant with that and could silently drift out of sync with it; the
// completeness/collision check that matters lives at the whole-catalog level
// in generate-contract-support.mjs's renderPublicContract, which fails
// generation if any catalog component's expected export goes missing.
const canonicalTypeName = (name) => name;

export default {
  input: join(moduleRoot, "generated", `types-${typesVersion}.json`),
  output: join(packageRoot, "src", "internal", "generated", "contract"),
  plugins: [
    {
      name: "zod",
      definitions: {
        case: "preserve",
        name: (name) => `z${canonicalTypeName(name)}`,
        types: {
          infer: {
            case: "preserve",
            name: canonicalTypeName,
          },
        },
      },
      $resolvers: {
        string(context) {
          if (context.schema.pattern !== "^[0-9a-f]{64}$") return;
          if (!isCanonicalProjectHashSchema(context.path)) return;
          return context.nodes.base(context)
            .attr("regex").call(context.$.expr("/^[0-9a-f]{64}$/"))
            .attr("brand").call().generic(context.$.type.literal("ProjectHash"));
        },
      },
      requests: false,
      responses: false,
    },
  ],
};

function requiredMatch(source, pattern, field, location) {
  const match = pattern.exec(source);
  if (match?.[1] !== undefined) return match[1];
  throw new Error(`TypeScript contract generation could not read ${field} from ${location}; the generator cannot identify its OpenAPI input; keep the canonical Go constant in the expected assignment form or update typescript/openapi-ts.config.mjs.`);
}

// isCanonicalProjectHashSchema anchors the ProjectHash brand to the schema's canonical
// $ref location (#/components/schemas/ProjectHash) rather than its pattern string alone.
// A pattern-only match would silently brand any future unrelated 64-lowercase-hex field
// (for example a SHA-256 content hash) as ProjectHash; matching on JSON-pointer identity
// keeps the nominal brand tied to the one real ProjectHash schema component, mirroring the
// $ref-anchored mechanism this generator's predecessor used before the Hey API rewrite.
function isCanonicalProjectHashSchema(path) {
  const segments = path?.["~ref"];
  return Array.isArray(segments) && segments.length === 3 && segments[0] === "components" && segments[1] === "schemas" && segments[2] === "ProjectHash";
}

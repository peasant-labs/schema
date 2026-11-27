import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

import { shouldBrandProjectHash } from "./scripts/lib/project-hash-resolver.mjs";

const packageRoot = dirname(fileURLToPath(import.meta.url));
const moduleRoot = join(packageRoot, "..");

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
  input: join(moduleRoot, "generated", "types-next.json"),
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
          if (!shouldBrandProjectHash(context.schema, context.path)) return;
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

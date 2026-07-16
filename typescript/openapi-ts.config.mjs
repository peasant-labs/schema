import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const packageRoot = dirname(fileURLToPath(import.meta.url));
const moduleRoot = join(packageRoot, "..");
const versionsSource = readFileSync(join(moduleRoot, "versions.go"), "utf8");
const typesVersion = requiredMatch(versionsSource, /TypesVersion\s*=\s*"([^"]+)"/, "TypesVersion", "versions.go");

const canonicalTypeName = (name) => ({
  cliloginquery: "CLILoginQuery",
  modelid: "ModelID",
  sessionid: "SessionID",
  transcriptid: "TranscriptID",
}[name.replace(/[^A-Za-z]/g, "").toLowerCase()] ?? name);

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

export function renderEnumInventoryDoc(enumCase) {
  if (enumCase.all_values.length >= enumCase.members.length) return "";
  const text = enumCase.subset_doc ?? `${enumCase.all_name} is a subset of ${enumCase.name}, not the full\nknown set. Every ${enumCase.name} member remains individually valid and accepted by is${enumCase.name}.`;
  if (typeof text !== "string" || text.trim() === "" || text.includes("*/")) {
    throw new Error(`TypeScript enum generation found an invalid subset_doc for ${enumCase.name} in enums.yaml; supply a nonempty comment without a comment terminator.`);
  }
  return `/**\n${text.split("\n").map((line) => ` * ${line}`).join("\n")}\n */\n`;
}

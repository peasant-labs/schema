export function applyCollectiveZodRefinements(source, spec) {
  let output = source;
  // The pinned generator emits date-time patterns but does not implement the
  // JSON Schema `not` zero-instant constraint. Apply that exact source pattern
  // only on the evidence fields carrying it, before cross-field refinements.
  for (const [name, component] of Object.entries(spec.components.schemas)) {
    for (const [field, property] of Object.entries(component.properties ?? {})) {
      if (property.format !== "date-time" || property.not?.pattern === undefined) continue;
      const start = output.indexOf(`export const z${name} = `), end = output.indexOf(`\n\nexport type ${name} =`, start);
      if (start < 0 || end < start) throw new Error(`Lossless timestamp generation lost ${name}; restore the evidence validator before publishing.`);
      const marker = `${field}: z.iso.datetime()`;
      const declaration = output.slice(start, end);
      if (!declaration.includes(marker)) throw new Error(`Lossless timestamp generation lost ${name}.${field}; inspect the pinned generator before publishing.`);
      const replacement = `${marker}.refine((value) => new RegExp(${JSON.stringify(property.pattern)}).test(value) && !new RegExp(${JSON.stringify(property.not.pattern)}).test(value), { message: "Collective timestamp requires nonzero UTC RFC3339 with explicit seconds and at most nine fractional digits; unsupported precision loses evidence, so correct the source timestamp ending in Z before retrying" })`;
      output = output.slice(0, start) + declaration.replace(marker, replacement) + output.slice(end);
    }
  }
  const names = ["CollectiveObservationScope", "ContributionTraversalRevision", "OperationIssue", "PullAssociationMetadataResult", "LocalConfirmedContributionReceipt", "LocalOperationDiagnostic", "CollectiveObservation", "CollectiveFetchDomainSummary", "CollectiveFetchDomainResult", "CollectiveFetchReport", "CollectiveObservationResponse", "CollectiveFetchSummary", "LocalContributionHistoryResponse", "VillageOrganizationCollectivesResponse", "VillageContributedCollectivesResponse"];
  for (const name of names) {
    const start = output.indexOf(`export const z${name} = `), end = output.indexOf(`\n\nexport type ${name} =`, start);
    if (start < 0 || end < start) throw new Error(`Collective contract generation lost ${name}; no runtime validator was produced; update the pinned generator postprocessor before publishing.`);
    let declaration = output.slice(start, end);
    if (name === "LocalContributionHistoryResponse") declaration = declaration.replaceAll("z.object({", "z.strictObject({");
    const refinement = `.superRefine((value, ctx) => { const issue = collectiveContractIssue("${name}", value); if (issue !== undefined) ctx.addIssue({ code: "custom", message: "Collective contract validation: " + issue + "; retain prior evidence and fetch validated metadata before retrying" }); })`;
    output = output.slice(0, start) + declaration.slice(0, -1) + refinement + ";" + output.slice(end);
  }
  return `import { collectiveContractIssue } from "../../../collective-validation.js";\n` + output;
}

export function applyCollectiveZodRefinements(source) {
  let output = source;
  const names = ["CollectiveObservationScope", "ContributionTraversalRevision", "OperationIssue", "PullAssociationMetadataResult", "LocalConfirmedContributionReceipt", "LocalOperationDiagnostic", "CollectiveObservation", "CollectiveFetchDomainSummary", "CollectiveFetchDomainResult", "CollectiveFetchReport", "CollectiveObservationResponse", "CollectiveFetchSummary", "LocalContributionHistoryResponse", "VillageOrganizationCollectivesResponse"];
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

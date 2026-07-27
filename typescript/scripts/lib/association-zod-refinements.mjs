const rawEvidence = "    evidence: z.array(zAssociationEvidenceObservation),";
const refinedEvidence = "    evidence: z.array(zAssociationEvidenceObservation).min(1),";
const rawSessionID = "    sessionId: zSessionID";
const refinedSessionID = "    sessionId: z.string().min(1)";

export function applyAssociationZodRefinements(source) {
  const declarations = {
    zAssociationEvidenceObservation: uniqueObjectDeclaration(source, "zAssociationEvidenceObservation"),
    zSessionAssociation: uniqueObjectDeclaration(source, "zSessionAssociation"),
    zGitContext: uniqueObjectDeclaration(source, "zGitContext"),
    zAnnotationPushItem: uniqueObjectDeclaration(source, "zAnnotationPushItem"),
    zAnnotationSummary: uniqueObjectDeclaration(source, "zAnnotationSummary"),
  };
  const refinements = {
    zAssociationEvidenceObservation: associationEvidenceObservationRefinement(),
    zSessionAssociation: sessionAssociationRefinement(),
    zGitContext: gitContextPublishedAssociationRefinement(),
    zAnnotationPushItem: annotationPushItemRefinement(),
    zAnnotationSummary: annotationSummaryRefinement(),
  };
  const signals = [
    propertySignal("zSessionAssociation.evidence", declarations.zSessionAssociation, rawEvidence, refinedEvidence),
    propertySignal("zSessionAssociation.sessionId", declarations.zSessionAssociation, rawSessionID, refinedSessionID),
    refinementSignal("zAssociationEvidenceObservation", declarations.zAssociationEvidenceObservation, refinements.zAssociationEvidenceObservation),
    refinementSignal("zSessionAssociation", declarations.zSessionAssociation, refinements.zSessionAssociation),
    refinementSignal("zGitContext", declarations.zGitContext, refinements.zGitContext),
    refinementSignal("zAnnotationPushItem", declarations.zAnnotationPushItem, refinements.zAnnotationPushItem),
    refinementSignal("zAnnotationSummary", declarations.zAnnotationSummary, refinements.zAnnotationSummary),
  ];

  if (signals.every((signal) => signal.state === "raw")) {
    return refineRawSource(source, declarations, refinements);
  }
  if (signals.every((signal) => signal.state === "refined")) return source;

  throw new Error(refinementStateError(signals));
}

function refineRawSource(source, declarations, refinements) {
  const sessionAssociation = replaceSingleText(
    replaceSingleText(declarations.zSessionAssociation.text, rawEvidence, refinedEvidence),
    rawSessionID,
    refinedSessionID,
  );
  return replaceDeclarations(source, [
    refinedDeclaration(declarations.zAssociationEvidenceObservation, refinements.zAssociationEvidenceObservation),
    refinedDeclaration({ ...declarations.zSessionAssociation, text: sessionAssociation }, refinements.zSessionAssociation),
    refinedDeclaration(declarations.zGitContext, refinements.zGitContext),
    refinedDeclaration(declarations.zAnnotationPushItem, refinements.zAnnotationPushItem),
    refinedDeclaration(declarations.zAnnotationSummary, refinements.zAnnotationSummary),
  ]);
}

function uniqueObjectDeclaration(source, schemaName) {
  const marker = `export const ${schemaName} = z.object({`;
  const starts = allIndexes(source, marker);
  if (starts.length !== 1) {
    return invalidDeclaration(schemaName, `found ${starts.length} declaration starts, expected exactly one`);
  }
  const start = starts[0];
  const end = source.indexOf("\n});", start);
  if (end === -1) {
    return invalidDeclaration(schemaName, "could not find the declaration terminator");
  }
  return { schemaName, start, end: end + "\n});".length, text: source.slice(start, end + "\n});".length) };
}

function invalidDeclaration(schemaName, reason) {
  return { schemaName, reason };
}

function propertySignal(name, declaration, raw, refined) {
  if (declaration.reason !== undefined) {
    return invalidSignal(name, `declaration unavailable: ${declaration.reason}`);
  }
  const rawCount = countText(declaration.text, raw);
  const refinedCount = countText(declaration.text, refined);
  if (rawCount === 1 && refinedCount === 0) return { name, state: "raw" };
  if (rawCount === 0 && refinedCount === 1) return { name, state: "refined" };
  return invalidSignal(name, `raw=${rawCount}, refined=${refinedCount}`);
}

function refinementSignal(name, declaration, refinement) {
  if (declaration.reason !== undefined) {
    return invalidSignal(name, `declaration unavailable: ${declaration.reason}`);
  }
  const rawCount = Number(isRawObjectDeclaration(declaration, name));
  const refinementCount = countText(declaration.text, refinement);
  const refined = refinementCount === 1 && isRefinedObjectDeclaration(declaration, name, refinement);
  if (rawCount === 1 && refinementCount === 0) return { name, state: "raw" };
  if (refined) return { name, state: "refined" };
  return invalidSignal(name, `raw=${rawCount}, refined=${refinementCount}, exactRefinement=${refined}`);
}

function isRawObjectDeclaration(declaration, schemaName) {
  return declaration.text.startsWith(`export const ${schemaName} = z.object({`)
    && declaration.text.endsWith("\n});")
    && countText(declaration.text, ".superRefine(") === 0;
}

function isRefinedObjectDeclaration(declaration, schemaName, refinement) {
  const suffix = `${refinement};`;
  if (!declaration.text.endsWith(suffix)) return false;
  const rawCandidate = {
    ...declaration,
    text: `${declaration.text.slice(0, -suffix.length)};`,
  };
  return isRawObjectDeclaration(rawCandidate, schemaName);
}

function refinedDeclaration(declaration, refinement) {
  return {
    ...declaration,
    replacement: `${declaration.text.slice(0, -1)}${refinement};`,
  };
}

function replaceDeclarations(source, declarations) {
  return [...declarations]
    .sort((left, right) => right.start - left.start)
    .reduce((result, declaration) => (
      `${result.slice(0, declaration.start)}${declaration.replacement}${result.slice(declaration.end)}`
    ), source);
}

function replaceSingleText(source, raw, refined) {
  const index = source.indexOf(raw);
  return `${source.slice(0, index)}${refined}${source.slice(index + raw.length)}`;
}

function allIndexes(source, text) {
  const indexes = [];
  let index = source.indexOf(text);
  while (index !== -1) {
    indexes.push(index);
    index = source.indexOf(text, index + text.length);
  }
  return indexes;
}

function countText(source, text) {
  return allIndexes(source, text).length;
}

function invalidSignal(name, detail) {
  return { name, state: "invalid", detail };
}

function refinementStateError(signals) {
  const observed = signals.map((signal) => `${signal.name}=${signal.state}${signal.detail === undefined ? "" : `(${signal.detail})`}`).join("; ");
  return `TypeScript root Zod generation/refinement operation in typescript/scripts/lib/association-zod-refinements.mjs rejected a partial, mixed, duplicate, missing, or drifted declaration state; observed signals: ${observed}; the pinned Hey API output or generated zod.gen.ts file drifted, so callers have no trustworthy root Zod output; inspect the generated source, update this postprocessor for the pinned Hey API shape, then regenerate.`;
}

function associationEvidenceObservationRefinement() {
  return `.superRefine((value, context) => {
    const isAbsent = (detail: unknown) => detail === undefined || detail === null;
    const isGoWhitespace = (codePoint: number) => (codePoint >= 0x0009 && codePoint <= 0x000D)
        || codePoint === 0x0020
        || codePoint === 0x0085
        || codePoint === 0x00A0
        || codePoint === 0x1680
        || (codePoint >= 0x2000 && codePoint <= 0x200A)
        || codePoint === 0x2028
        || codePoint === 0x2029
        || codePoint === 0x202F
        || codePoint === 0x205F
        || codePoint === 0x3000;
    const hasNonEmptyString = (detail: unknown) => {
        if (typeof detail !== "string") return false;
        for (const character of detail) {
            const codePoint = character.codePointAt(0);
            if (codePoint !== undefined && !isGoWhitespace(codePoint)) return true;
        }
        return false;
    };
    const hasValidTouchedFilePath = (detail: unknown) => typeof detail === "string"
        && detail !== ""
        && !detail.startsWith("/")
        && !detail.startsWith("\\\\")
        && !detail.includes("\\\\")
        && !/^[A-Za-z]:\\//.test(detail)
        && !detail.split("/").some((segment) => segment === "" || segment === "." || segment === "..");
    const addIssue = (message: string) => context.addIssue({ code: "custom", message });
    switch (value.kind) {
        case "recorded_commit":
            if (!hasNonEmptyString(value.recordedCommitHash)) addIssue("recorded_commit evidence requires a non-empty recordedCommitHash");
            if (!isAbsent(value.touchedFilePath) || !isAbsent(value.branchName) || !isAbsent(value.windowStartMs) || !isAbsent(value.windowEndMs)) addIssue("recorded_commit evidence must not populate another detail arm");
            return;
        case "touched_file":
            if (!hasValidTouchedFilePath(value.touchedFilePath)) addIssue("touched_file evidence requires a non-empty repository-relative touchedFilePath");
            if (!isAbsent(value.recordedCommitHash) || !isAbsent(value.branchName) || !isAbsent(value.windowStartMs) || !isAbsent(value.windowEndMs)) addIssue("touched_file evidence must not populate another detail arm");
            return;
        case "branch_membership":
            if (!hasNonEmptyString(value.branchName)) addIssue("branch_membership evidence requires a non-empty branchName");
            if (!isAbsent(value.recordedCommitHash) || !isAbsent(value.touchedFilePath) || !isAbsent(value.windowStartMs) || !isAbsent(value.windowEndMs)) addIssue("branch_membership evidence must not populate another detail arm");
            return;
        case "time_window":
            if (value.windowStartMs === undefined || value.windowStartMs === null || value.windowEndMs === undefined || value.windowEndMs === null) {
                addIssue("time_window evidence requires windowStartMs and windowEndMs");
            } else if (value.windowStartMs > value.windowEndMs) {
                addIssue("time_window evidence requires windowStartMs less than or equal to windowEndMs");
            }
            if (!isAbsent(value.recordedCommitHash) || !isAbsent(value.touchedFilePath) || !isAbsent(value.branchName)) addIssue("time_window evidence must not populate another detail arm");
            return;
    }
})`;
}

function sessionAssociationRefinement() {
  return `.superRefine((value, context) => {
    const kindOrder = ["recorded_commit", "touched_file", "branch_membership", "time_window"];
    const compareStrings = (left: string, right: string) => {
        const leftBytes = new TextEncoder().encode(left);
        const rightBytes = new TextEncoder().encode(right);
        const sharedLength = Math.min(leftBytes.length, rightBytes.length);
        for (let index = 0; index < sharedLength; index += 1) {
            const leftByte = leftBytes[index]!;
            const rightByte = rightBytes[index]!;
            if (leftByte < rightByte) return -1;
            if (leftByte > rightByte) return 1;
        }
        return compareNumbers(leftBytes.length, rightBytes.length);
    };
    const compareNumbers = (left: number, right: number) => left < right ? -1 : left > right ? 1 : 0;
    const compareObservations = (left: AssociationEvidenceObservation, right: AssociationEvidenceObservation): number => {
        const leftKindOrder = kindOrder.indexOf(left.kind);
        const rightKindOrder = kindOrder.indexOf(right.kind);
        if (leftKindOrder !== rightKindOrder) return compareNumbers(leftKindOrder, rightKindOrder);
        switch (left.kind) {
            case "recorded_commit": return compareStrings(typeof left.recordedCommitHash === "string" ? left.recordedCommitHash : "", typeof right.recordedCommitHash === "string" ? right.recordedCommitHash : "");
            case "touched_file": return compareStrings(typeof left.touchedFilePath === "string" ? left.touchedFilePath : "", typeof right.touchedFilePath === "string" ? right.touchedFilePath : "");
            case "branch_membership": return compareStrings(typeof left.branchName === "string" ? left.branchName : "", typeof right.branchName === "string" ? right.branchName : "");
            case "time_window": {
                const startOrder = compareNumbers(left.windowStartMs ?? 0, right.windowStartMs ?? 0);
                return startOrder === 0 ? compareNumbers(left.windowEndMs ?? 0, right.windowEndMs ?? 0) : startOrder;
            }
        }
        return 0;
    };
    for (let index = 1; index < value.evidence.length; index += 1) {
        const previous = value.evidence[index - 1];
        const current = value.evidence[index];
        if (previous === undefined || current === undefined) continue;
        if (compareObservations(previous, current) >= 0) {
            context.addIssue({ code: "custom", message: "association evidence must be duplicate-free and in canonical kind/detail order" });
        }
    }
})`;
}

function annotationSummaryRefinement() {
  return `.superRefine((value, context) => {
    const isPresent = (detail: unknown) => detail !== undefined && detail !== null;
    const hasNonEmptyString = (detail: unknown) => typeof detail === "string" && detail !== "";
    const hasOnly = (allowed: readonly string[]) => {
        const targetFields: readonly [string, unknown][] = [
            ["targetAssociationId", value.targetAssociationId],
            ["targetSessionId", value.targetSessionId],
            ["targetEntryIndex", value.targetEntryIndex],
            ["targetEntryEndIndex", value.targetEntryEndIndex],
            ["targetAnnotationId", value.targetAnnotationId],
            ["targetProjectHash", value.targetProjectHash],
            ["targetFilePath", value.targetFilePath],
            ["targetContentHash", value.targetContentHash]
        ];
        return targetFields.every(([name, detail]) => allowed.includes(name) || !isPresent(detail));
    };
    const addIssue = (message: string) => context.addIssue({ code: "custom", message });
    switch (value.targetKind) {
        case "association":
            if (!isPresent(value.targetAssociationId)) addIssue("association annotations require targetAssociationId");
            if (!hasOnly(["targetAssociationId"])) addIssue("association annotations must not mix target arms");
            return;
        case "session":
            if (!hasNonEmptyString(value.targetSessionId)) addIssue("session annotations require a non-empty targetSessionId");
            if (!hasOnly(["targetSessionId"])) addIssue("session annotations must not mix target arms");
            return;
        case "entry":
            if (!hasNonEmptyString(value.targetSessionId) || !isPresent(value.targetEntryIndex)) addIssue("entry annotations require targetSessionId and targetEntryIndex");
            if (!hasOnly(["targetSessionId", "targetEntryIndex", "targetEntryEndIndex"])) addIssue("entry annotations must not mix target arms");
            if (value.targetEntryIndex !== undefined && value.targetEntryIndex !== null && value.targetEntryEndIndex !== undefined && value.targetEntryEndIndex !== null && value.targetEntryEndIndex <= value.targetEntryIndex) addIssue("entry annotation targetEntryEndIndex must be greater than targetEntryIndex");
            return;
        case "annotation":
            if (!hasNonEmptyString(value.targetAnnotationId)) addIssue("annotation targets require a non-empty targetAnnotationId");
            if (!hasOnly(["targetAnnotationId"])) addIssue("annotation targets must not mix target arms");
            return;
        case "project":
            if (!isPresent(value.targetProjectHash)) addIssue("project annotations require targetProjectHash");
            if (!hasOnly(["targetProjectHash"])) addIssue("project annotations must not mix target arms");
            return;
        case "file_version":
            if (!hasNonEmptyString(value.targetFilePath) || !hasNonEmptyString(value.targetContentHash)) addIssue("file_version annotations require non-empty targetFilePath and targetContentHash");
            if (!hasOnly(["targetFilePath", "targetContentHash"])) addIssue("file_version annotations must not mix target arms");
            return;
    }
})`;
}

function annotationPushItemRefinement() {
  return `.superRefine((value, context) => {
    const isPresent = (detail: unknown) => detail !== undefined && detail !== null;
    const hasNonEmptyString = (detail: unknown) => typeof detail === "string" && detail !== "";
    const hasOnly = (allowed: readonly string[]) => {
        const targetFields: readonly [string, unknown][] = [
            ["targetAssociationId", value.targetAssociationId],
            ["sessionId", value.sessionId],
            ["entryTarget", value.entryTarget],
            ["annotationId", value.annotationId],
            ["projectHash", value.projectHash]
        ];
        return targetFields.every(([name, detail]) => allowed.includes(name) || !isPresent(detail));
    };
    const addIssue = (message: string) => context.addIssue({ code: "custom", message });
    switch (value.targetKind) {
        case "association":
            if (!isPresent(value.targetAssociationId)) addIssue("association annotations require targetAssociationId");
            if (!hasOnly(["targetAssociationId"])) addIssue("association annotations must not mix target arms");
            return;
        case "session":
            if (!hasNonEmptyString(value.sessionId)) addIssue("session annotations require a non-empty sessionId");
            if (!hasOnly(["sessionId"])) addIssue("session annotations must not mix target arms");
            return;
        case "entry":
            if (!isPresent(value.entryTarget)) addIssue("entry annotations require entryTarget");
            if (!hasOnly(["entryTarget"])) addIssue("entry annotations must not mix target arms");
            return;
        case "annotation":
            if (!hasNonEmptyString(value.annotationId)) addIssue("annotation targets require a non-empty annotationId");
            if (!hasOnly(["annotationId"])) addIssue("annotation targets must not mix target arms");
            return;
        case "project":
            if (!isPresent(value.projectHash)) addIssue("project annotations require projectHash");
            if (!hasOnly(["projectHash"])) addIssue("project annotations must not mix target arms");
            return;
        case "file_version":
            addIssue("file_version annotations have no AnnotationPushItem target representation");
            return;
    }
})`;
}

function gitContextPublishedAssociationRefinement() {
  return `.superRefine((value, context) => {
    if (value.associations === undefined || value.associations === null) return;
    const seenIDs = new Map<string, number>();
    const seenObservedHashes = new Map<string, number>();
    for (const [index, association] of value.associations.entries()) {
        if (association.observedCommitHash.trim() === "") {
            context.addIssue({ code: "custom", path: ["associations", index, "observedCommitHash"], message: "published associations require a non-empty observedCommitHash" });
        }
        const priorID = seenIDs.get(association.id);
        if (priorID !== undefined) {
            context.addIssue({ code: "custom", path: ["associations", index, "id"], message: "published association IDs must be unique within one request" });
        } else {
            seenIDs.set(association.id, index);
        }
        const priorObservedHash = seenObservedHashes.get(association.observedCommitHash);
        if (priorObservedHash !== undefined) {
            context.addIssue({ code: "custom", path: ["associations", index, "observedCommitHash"], message: "published association observedCommitHash values must be unique within one request" });
        } else {
            seenObservedHashes.set(association.observedCommitHash, index);
        }
    }
})`;
}

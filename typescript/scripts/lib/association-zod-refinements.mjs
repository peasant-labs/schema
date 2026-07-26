export function applyAssociationZodRefinements(source) {
  let refined = source;
  refined = replaceExactly(refined, /    evidence: z\.array\(zAssociationEvidenceObservation\),/, "    evidence: z.array(zAssociationEvidenceObservation).min(1),", "zSessionAssociation.evidence");
  refined = replaceSessionAssociationSessionID(refined);
  refined = appendRefinement(refined, "zAssociationEvidenceObservation", associationEvidenceObservationRefinement());
  refined = appendRefinement(refined, "zSessionAssociation", sessionAssociationRefinement());
  refined = appendRefinement(refined, "zAnnotationSummary", annotationSummaryRefinement());
  return refined;
}

function replaceExactly(source, pattern, replacement, name) {
  const matches = source.match(new RegExp(pattern, "g")) ?? [];
  if (matches.length !== 1) {
    throw new Error(`TypeScript Zod refinement generation could not locate exactly one ${name} declaration; the pinned Hey API output shape changed, so association validation was not emitted; inspect the generator output and update typescript/scripts/lib/association-zod-refinements.mjs before publishing.`);
  }
  return source.replace(pattern, replacement);
}

function replaceSessionAssociationSessionID(source) {
  const pattern = /(export const zSessionAssociation = z\.object\(\{[\s\S]*?\n    sessionId: )zSessionID(\n\}\);)/g;
  const matches = source.match(pattern) ?? [];
  if (matches.length !== 1) {
    throw new Error("TypeScript Zod refinement generation could not locate exactly one zSessionAssociation.sessionId declaration; the pinned Hey API output shape changed, so Go-compatible association session IDs were not emitted; inspect the generator output and update typescript/scripts/lib/association-zod-refinements.mjs before publishing.");
  }
  return source.replace(pattern, "$1z.string().min(1)$2");
}

function appendRefinement(source, schemaName, refinement) {
  const pattern = new RegExp(`export const ${schemaName} = z\\.object\\(\\{[\\s\\S]*?\\n\\}\\);`, "g");
  const matches = source.match(pattern) ?? [];
  if (matches.length !== 1) {
    throw new Error(`TypeScript Zod refinement generation could not locate exactly one ${schemaName} object declaration; the pinned Hey API output shape changed, so the root runtime contract would lose its structural validation; inspect the generator output and update typescript/scripts/lib/association-zod-refinements.mjs before publishing.`);
  }
  const declaration = matches[0];
  return source.replace(pattern, `${declaration.slice(0, -1)}${refinement};`);
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

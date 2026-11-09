// The generated zTurnContentPart is a permissive object schema: every field is
// optional except type. That shape alone accepts a text part with an image, an
// image part with no image, a published image carrying a local ref, a local
// inline image with no ref or size, a non-recorded media type, and an
// unsupported image with stored bytes. The Go validator refuses every one of
// those shapes, so this refinement restores the part-level parity. The caller's
// mode (local detail versus published detail) is enforced separately by the
// runtime parsers because one shared schema cannot know which payload mode is
// being decoded.
const declarationPattern = /export const zTurnContentPart = z\.object\(\{[\s\S]*?\}\);/;

export function applyContentPartsZodRefinements(source) {
  const original = source.match(declarationPattern)?.[0];
  if (original === undefined) {
    throw new Error("TypeScript contract generation could not apply content-part validation to zTurnContentPart in src/internal/generated/contract/zod.gen.ts; the generated declaration no longer matches `z.object({...})`, so the package would accept invalid text/image part shapes the Go contract rejects; inspect the pinned generator output and update scripts/lib/content-parts-zod-refinements.mjs.");
  }
  const refinement = `.superRefine((part, ctx) => {
    const issue = (message: string) => ctx.addIssue({ code: "custom", message });
    if (part.type === "text") {
      if (part.image !== undefined) issue("a text part must not carry image");
      return;
    }
    if (part.image === undefined) {
      issue("an image part requires image");
      return;
    }
    if (part.text !== undefined) issue("an image part must not carry text");
    const image = part.image;
    if (image.byteSize !== undefined && image.byteSize < 0) issue("byteSize must be nonnegative");
    if (image.width !== undefined && image.width < 0) issue("width must be nonnegative");
    if (image.height !== undefined && image.height < 0) issue("height must be nonnegative");
    if (image.ref !== undefined && !/^[0-9a-f]{64}$/.test(image.ref)) issue("ref must be lowercase 64-hex");
    if (image.omittedReason === "not_published") {
      if (image.ref !== undefined || image.mediaType !== undefined || image.byteSize !== undefined || image.width !== undefined || image.height !== undefined) issue("a published image must not carry local fields");
      return;
    }
    if (image.omittedReason === "unsupported") {
      if (image.mediaType !== undefined) issue("an unsupported image must not carry mediaType");
      if (image.ref !== undefined || image.byteSize !== undefined) issue("an unsupported image must not carry ref or byteSize");
      return;
    }
    if (image.mediaType !== "image/png" && image.mediaType !== "image/jpeg") issue("a local image requires image/png or image/jpeg");
    if (image.omittedReason === "too_large") {
      if (image.ref !== undefined) issue("a too_large image must not carry ref");
      if (image.byteSize !== undefined) issue("a too_large image must not carry byteSize");
      return;
    }
    if (image.omittedReason === "missing") {
      if (image.byteSize !== undefined) issue("a missing image must not carry byteSize");
      return;
    }
    if (image.ref === undefined) issue("an inline image requires ref");
    if (image.byteSize === undefined || image.byteSize === 0) issue("an inline image requires a positive byteSize");
  })`;
  return source.replace(original, `${original.slice(0, -1)}${refinement};`);
}

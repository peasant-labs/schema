package schema

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// ImagePartMode selects which image-part shape a caller is validating. Local is
// the durable/local-detail shape: an accepted image names its content reference
// and byte size. Published is the bounded publication shape: a published
// transcript never carries bytes, a reference, media type, size, or dimensions,
// and instead says the image was not published. The same ImagePart struct
// carries both shapes so one type describes the ordered content list end to
// end.
type ImagePartMode string

const (
	// ImagePartModeLocal validates the durable local-detail image shape.
	ImagePartModeLocal ImagePartMode = "local"
	// ImagePartModePublished validates the publication image shape.
	ImagePartModePublished ImagePartMode = "published"

	// imagePartModeEnvelope accepts either released shape at the transcript
	// envelope boundary. TranscriptContent carries both a standalone local
	// export and the published wire, so that one decoder validates a part that
	// is valid in either mode. It is deliberately unexported: producers pick a
	// real mode when they build or publish content.
	imagePartModeEnvelope ImagePartMode = "envelope"
)

// ImageMediaTypePNG and ImageMediaTypeJPEG are the only recorded image media
// types. An image whose declared type is outside this set is recorded as an
// unsupported omission instead.
const (
	ImageMediaTypePNG  = "image/png"
	ImageMediaTypeJPEG = "image/jpeg"
)

// AllImageMediaTypes is the canonical closed list of recorded image media
// types.
var AllImageMediaTypes = []string{ImageMediaTypePNG, ImageMediaTypeJPEG}

// OmittedReason says why an image part carries no bytes. The empty value means
// the image bytes are present locally. not_published means the image was
// deliberately omitted from a published transcript.
type OmittedReason string

const (
	// OmittedReasonTooLarge records an image above the accepted size limit. The
	// part keeps its declared media type and dimensions when known but carries
	// no reference or bytes.
	OmittedReasonTooLarge OmittedReason = "too_large"
	// OmittedReasonUnsupported records a media type this contract does not
	// accept (for example webp or gif). The part carries no reference, bytes,
	// or media type because the unsupported spelling is outside the closed set.
	OmittedReasonUnsupported OmittedReason = "unsupported"
	// OmittedReasonMissing records a recorded image whose referenced file is
	// absent. The reference may still be known; the size is always zero.
	OmittedReasonMissing OmittedReason = "missing"
	// OmittedReasonNotPublished records an image the publisher deliberately
	// omitted. It is the only reason a published transcript may carry.
	OmittedReasonNotPublished OmittedReason = "not_published"
)

// AllOmittedReasons is the canonical closed list of image omission reasons.
var AllOmittedReasons = []OmittedReason{
	OmittedReasonTooLarge, OmittedReasonUnsupported, OmittedReasonMissing, OmittedReasonNotPublished,
}

// IsValid reports whether the reason is a published omission reason.
func (r OmittedReason) IsValid() bool {
	switch r {
	case OmittedReasonTooLarge, OmittedReasonUnsupported, OmittedReasonMissing, OmittedReasonNotPublished:
		return true
	}
	return false
}

func (r OmittedReason) String() string { return string(r) }

// JSONSchema implements jsonschema.Exposer.
func (OmittedReason) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Omitted Reason",
		"Why an image part carries no bytes: too_large, unsupported, missing, or not_published",
		AllOmittedReasons,
	), nil
}

// TurnPartType is the closed kind set of one ordered turn content part.
type TurnPartType string

const (
	// TurnPartTypeText is an ordered text block of a turn.
	TurnPartTypeText TurnPartType = "text"
	// TurnPartTypeImage is an ordered image block of a turn.
	TurnPartTypeImage TurnPartType = "image"
)

// AllTurnPartTypes is the canonical closed list of turn content part kinds.
var AllTurnPartTypes = []TurnPartType{TurnPartTypeText, TurnPartTypeImage}

// IsValid reports whether the part type is a published kind.
func (t TurnPartType) IsValid() bool {
	switch t {
	case TurnPartTypeText, TurnPartTypeImage:
		return true
	}
	return false
}

func (t TurnPartType) String() string { return string(t) }

// JSONSchema implements jsonschema.Exposer.
func (TurnPartType) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Turn Part Type",
		"Kind of one ordered turn content part: text or image",
		AllTurnPartTypes,
	), nil
}

// ImagePart is one ordered image block of a turn. The local shape names the
// content-addressed reference and the recorded metadata; the published shape
// carries only the omission reason, never bytes, a reference, a media type, a
// size, or dimensions.
type ImagePart struct {
	// Ref is the lowercase 64-hex content address of the stored image. It is
	// present for an inline local image and may be present for a missing local
	// image; it is always absent from a published transcript.
	Ref string `json:"ref,omitempty" pattern:"^[0-9a-f]{64}$" minLength:"64" maxLength:"64" description:"Lowercase 64-hex content address of the stored image. Absent when the bytes are omitted or the transcript is published."`
	// MediaType is the recorded image/png or image/jpeg type. It is absent for
	// an unsupported source type, which cannot be represented in this closed
	// set, and for a published transcript.
	MediaType string `json:"mediaType,omitempty" description:"Recorded image/png or image/jpeg media type. Absent for an unsupported source type or a published transcript."`
	// ByteSize is the exact stored byte size. It is zero when no bytes are
	// carried and always absent from a published transcript.
	ByteSize int `json:"byteSize,omitempty" minimum:"0" description:"Exact stored byte size. Zero when the bytes are omitted; always zero or absent for a published transcript."`
	// Width and Height are the optional recorded pixel dimensions. Zero means
	// unknown; a published transcript carries neither.
	Width  int `json:"width,omitempty" minimum:"0" description:"Optional recorded pixel width; zero means unknown. Absent from a published transcript."`
	Height int `json:"height,omitempty" minimum:"0" description:"Optional recorded pixel height; zero means unknown. Absent from a published transcript."`
	// OmittedReason says why no bytes are carried. The empty value means local
	// bytes are present.
	OmittedReason OmittedReason `json:"omittedReason,omitempty" description:"Why the bytes are absent: too_large, unsupported, missing, or not_published. Omit when local bytes are present."`
}

// TurnContentPart is one ordered content block of a turn. Parts preserve source
// order and replace no existing field: Content remains the concatenated text
// fallback and ToolCalls remains tool evidence.
type TurnContentPart struct {
	Type TurnPartType `json:"type"`
	// Text is the ordered text block. It is set only for a text part.
	Text string `json:"text,omitempty"`
	// Image is the ordered image block. It is set only for an image part.
	Image *ImagePart `json:"image,omitempty" nullable:"false"`
}

// ValidateTurnContentPart validates one ordered part for the given mode. A text
// part is accepted in both modes. An image part delegates to
// ValidateImagePart.
func ValidateTurnContentPart(mode ImagePartMode, part TurnContentPart) error {
	if mode != ImagePartModeLocal && mode != ImagePartModePublished {
		return fmt.Errorf(
			"turn content part validation failed at schema.ValidateTurnContentPart while selecting mode %q: the mode is not local or published, so the accepted shape is undefined; pass ImagePartModeLocal for durable detail or ImagePartModePublished for a published transcript",
			mode,
		)
	}
	if !part.Type.IsValid() {
		return fmt.Errorf(
			"turn content part validation failed at schema.ValidateTurnContentPart while reading part type %q: the type is outside the closed text|image set, so a consumer cannot choose a renderer; emit type text or type image from AllTurnPartTypes",
			part.Type,
		)
	}
	switch part.Type {
	case TurnPartTypeText:
		if part.Image != nil {
			return fmt.Errorf(
				"turn content part validation failed at schema.ValidateTurnContentPart while reading a text part: a text part carries an image object, so the part has two renderers; emit text without image",
			)
		}
		if !utf8.ValidString(part.Text) {
			return fmt.Errorf(
				"turn content part validation failed at schema.ValidateTurnContentPart while reading a text part: text is not valid UTF-8, so the ordered block would be replaced during encoding; emit a valid Unicode string",
			)
		}
		return nil
	case TurnPartTypeImage:
		if part.Image == nil {
			return fmt.Errorf(
				"turn content part validation failed at schema.ValidateTurnContentPart while reading an image part: type image has no image object, so the part has no evidence to render; emit image with a ref or an omission reason",
			)
		}
		if part.Text != "" {
			return fmt.Errorf(
				"turn content part validation failed at schema.ValidateTurnContentPart while reading an image part: type image also carries text, so the part has two renderers; emit the text as its own text part in source order",
			)
		}
		return ValidateImagePart(mode, *part.Image)
	default:
		return fmt.Errorf(
			"turn content part validation failed at schema.ValidateTurnContentPart while reading part type %q: the type is outside the closed text|image set, so a consumer cannot choose a renderer; emit type text or type image from AllTurnPartTypes",
			part.Type,
		)
	}
}

// ValidateImagePart validates one image part for the given mode. Local requires
// a known media type except for an unsupported source type, a lowercase 64-hex
// reference whenever the reason is not exempt, a zero size whenever the bytes
// are omitted, and the exact zero-or-present size rule. Published requires the
// not_published reason and forbids every local field. Width and height are
// always optional and nonnegative.
func ValidateImagePart(mode ImagePartMode, part ImagePart) error {
	if part.OmittedReason != "" && !part.OmittedReason.IsValid() {
		return fmt.Errorf(
			"image part validation failed at schema.ValidateImagePart while reading omittedReason %q: the reason is outside the closed too_large|unsupported|missing|not_published set, so a consumer cannot explain the missing bytes; use a member of AllOmittedReasons",
			part.OmittedReason,
		)
	}
	if part.ByteSize < 0 || part.Width < 0 || part.Height < 0 {
		return fmt.Errorf(
			"image part validation failed at schema.ValidateImagePart while reading byteSize=%d width=%d height=%d: a size or dimension is negative, so the metadata is not a measurement; emit nonnegative integers with zero meaning unknown",
			part.ByteSize, part.Width, part.Height,
		)
	}
	switch mode {
	case ImagePartModePublished:
		return validatePublishedImagePart(part)
	case ImagePartModeLocal:
		return validateLocalImagePart(part)
	default:
		return fmt.Errorf(
			"image part validation failed at schema.ValidateImagePart while selecting mode %q: the mode is not local or published, so the accepted shape is undefined; pass ImagePartModeLocal for durable detail or ImagePartModePublished for a published transcript",
			mode,
		)
	}
}

func validatePublishedImagePart(part ImagePart) error {
	if part.OmittedReason != OmittedReasonNotPublished {
		return fmt.Errorf(
			"published image part validation failed at schema.ValidateImagePart while reading omittedReason %q: a published transcript must say not_published, so a reader knows the bytes were withheld on purpose; set omittedReason to not_published",
			part.OmittedReason,
		)
	}
	if part.Ref != "" || part.MediaType != "" || part.ByteSize != 0 || part.Width != 0 || part.Height != 0 {
		return fmt.Errorf(
			"published image part validation failed at schema.ValidateImagePart while reading a not_published image: ref, mediaType, byteSize, width, or height is present, so the published wire would leak local image evidence or a byte count; publish only {type: image, omittedReason: not_published}",
		)
	}
	return nil
}

func validateLocalImagePart(part ImagePart) error {
	if part.OmittedReason == OmittedReasonNotPublished {
		return fmt.Errorf(
			"local image part validation failed at schema.ValidateImagePart while reading omittedReason not_published: a durable local image was never published, so the reason cannot describe this shape; use too_large, unsupported, missing, or omit the reason",
		)
	}
	if part.Ref != "" && !validImageRef(part.Ref) {
		return fmt.Errorf(
			"local image part validation failed at schema.ValidateImagePart while reading ref %q: the reference is not a lowercase 64-hex content address, so the attachment cannot be resolved; send the sha256 hex digest of the stored bytes or omit the reference",
			part.Ref,
		)
	}
	if part.OmittedReason == OmittedReasonUnsupported {
		if part.MediaType != "" {
			return fmt.Errorf(
				"local image part validation failed at schema.ValidateImagePart while reading an unsupported image: mediaType %q is present, but an unsupported source type is outside the recorded image/png|image/jpeg set; omit mediaType and keep the unsupported reason",
				part.MediaType,
			)
		}
	} else if !validImageMediaType(part.MediaType) {
		return fmt.Errorf(
			"local image part validation failed at schema.ValidateImagePart while reading mediaType %q: the media type is not image/png or image/jpeg, so a renderer cannot choose a decoder; record an accepted type or mark the image unsupported",
			part.MediaType,
		)
	}
	switch part.OmittedReason {
	case "":
		if part.Ref == "" {
			return fmt.Errorf(
				"local image part validation failed at schema.ValidateImagePart while reading an inline image: ref is absent with no omission reason, so the accepted bytes cannot be found; emit the 64-hex content address or set an omission reason",
			)
		}
		if part.ByteSize == 0 {
			return fmt.Errorf(
				"local image part validation failed at schema.ValidateImagePart while reading an inline image: byteSize is zero with no omission reason, so a stored image claims no bytes; record the exact stored size or set an omission reason",
			)
		}
	case OmittedReasonTooLarge:
		if part.Ref != "" {
			return fmt.Errorf(
				"local image part validation failed at schema.ValidateImagePart while reading a too_large image: ref %q is present, but a too-large image keeps no bytes; omit ref and keep only the media type and known dimensions",
				part.Ref,
			)
		}
		if part.ByteSize != 0 {
			return fmt.Errorf(
				"local image part validation failed at schema.ValidateImagePart while reading a too_large image: byteSize=%d is present, but a too-large image keeps no bytes; set byteSize to zero",
				part.ByteSize,
			)
		}
	case OmittedReasonUnsupported:
		if part.Ref != "" || part.ByteSize != 0 {
			return fmt.Errorf(
				"local image part validation failed at schema.ValidateImagePart while reading an unsupported image: ref or byteSize is present, but an unsupported source type is never stored; omit ref and byteSize",
			)
		}
	case OmittedReasonMissing:
		if part.ByteSize != 0 {
			return fmt.Errorf(
				"local image part validation failed at schema.ValidateImagePart while reading a missing image: byteSize=%d is present, but the file is absent so no size can be measured; set byteSize to zero",
				part.ByteSize,
			)
		}
	}
	return nil
}

func validImageRef(ref string) bool {
	if len(ref) != 64 || ref != strings.ToLower(ref) {
		return false
	}
	_, err := hex.DecodeString(ref)
	return err == nil
}

func validImageMediaType(mediaType string) bool {
	for _, accepted := range AllImageMediaTypes {
		if mediaType == accepted {
			return true
		}
	}
	return false
}

// validateTurnContentPartsAt validates ordered parts for one turn, adding the
// turn location to the error. The envelope mode accepts either the local or the
// published shape because TranscriptContent carries both a standalone local
// export and the published wire.
func validateTurnContentPartsAt(parts []TurnContentPart, mode ImagePartMode, path string, turnIndex int) error {
	for i, part := range parts {
		if mode != imagePartModeEnvelope {
			if err := ValidateTurnContentPart(mode, part); err != nil {
				return fmt.Errorf("session detail validation failed at schema.ValidateSessionDetailPayload for %s[%d].contentParts[%d]: %w", path, turnIndex, i, err)
			}
			continue
		}
		localErr := ValidateTurnContentPart(ImagePartModeLocal, part)
		publishedErr := ValidateTurnContentPart(ImagePartModePublished, part)
		if localErr != nil && publishedErr != nil {
			return fmt.Errorf(
				"session detail validation failed at schema.ValidateSessionDetailPayload for %s[%d].contentParts[%d]: the part is neither a valid local nor a valid published part, so an envelope reader cannot interpret it; local: %v; published: %v",
				path, turnIndex, i, localErr, publishedErr,
			)
		}
	}
	return nil
}

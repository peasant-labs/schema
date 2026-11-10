package schema

import (
	"fmt"
	"regexp"
)

// AttachmentRoutePath is the local content-addressed image read. The path
// parameter is the sha256 of the stored bytes, not a session identifier.
const AttachmentRoutePath = "/api/v1/attachments/{ref}"

// AttachmentRouteOperationID is the OpenAPI operationId of AttachmentRoutePath.
const AttachmentRouteOperationID = "getAttachment"

// AttachmentRefPattern is the path-parameter pattern for a stored image. It is
// the same lowercase 64-hex sha256 rule ImagePart.ref uses, so a renderer and
// the local route cannot disagree about which references resolve.
const AttachmentRefPattern = `^[0-9a-f]{64}$`

// AttachmentCacheControl is the Cache-Control value of a successful image read.
// The URL is the digest, so a 200 response never changes.
const AttachmentCacheControl = "public, max-age=31536000, immutable"

// Attachment route statuses. A malformed ref is refused before any file is
// opened. A well-formed ref with no row, or a row whose file is missing, is
// not found. A present png or jpeg is 200.
const (
	AttachmentRouteStatusOK        = 200
	AttachmentRouteStatusMalformed = 400
	AttachmentRouteStatusNotFound  = 404
)

// Machine-readable LocalErrorResponse codes for the attachment route. A 200
// response has no code.
const (
	AttachmentRouteCodeMalformed = "attachment_ref_malformed"
	AttachmentRouteCodeNotFound  = "attachment_not_found"
)

// AttachmentRouteOutcome is the closed set of results GET
// /api/v1/attachments/{ref} declares. missing_row and missing_file are
// distinct inputs that share one 404, so a caller cannot collapse a missing
// file into a malformed ref.
type AttachmentRouteOutcome string

const (
	AttachmentRouteOutcomePNG         AttachmentRouteOutcome = "png"
	AttachmentRouteOutcomeJPEG        AttachmentRouteOutcome = "jpeg"
	AttachmentRouteOutcomeMalformed   AttachmentRouteOutcome = "malformed"
	AttachmentRouteOutcomeMissingRow  AttachmentRouteOutcome = "missing_row"
	AttachmentRouteOutcomeMissingFile AttachmentRouteOutcome = "missing_file"
)

// AllAttachmentRouteOutcomes is the canonical closed list.
var AllAttachmentRouteOutcomes = []AttachmentRouteOutcome{
	AttachmentRouteOutcomePNG,
	AttachmentRouteOutcomeJPEG,
	AttachmentRouteOutcomeMalformed,
	AttachmentRouteOutcomeMissingRow,
	AttachmentRouteOutcomeMissingFile,
}

// IsValid reports whether the outcome belongs to the closed set.
func (o AttachmentRouteOutcome) IsValid() bool {
	for _, known := range AllAttachmentRouteOutcomes {
		if o == known {
			return true
		}
	}
	return false
}

func (o AttachmentRouteOutcome) String() string { return string(o) }

// AttachmentRouteResponse is one declared response of the local image read.
// MediaType and CacheControl are set only for 200. Code is set only for a
// refusal.
type AttachmentRouteResponse struct {
	Status       int
	MediaType    ImageMediaType
	CacheControl string
	Code         string
}

var attachmentRefPattern = regexp.MustCompile(AttachmentRefPattern)

// ValidateAttachmentRef accepts a lowercase 64-hex sha256 and rejects every
// other spelling. The local attachment route answers 400 for a rejection and
// does not open a file.
func ValidateAttachmentRef(ref string) error {
	if attachmentRefPattern.MatchString(ref) {
		return nil
	}
	return fmt.Errorf(
		"attachment route validation failed at schema.ValidateAttachmentRef while reading ref %q: the path parameter is not a lowercase 64-hex sha256, so GET %s refuses it with %d and does not open a file; pass the sha256 hex digest of the stored image",
		ref, AttachmentRoutePath, AttachmentRouteStatusMalformed,
	)
}

// AttachmentRouteResponseFor returns the declared response for one closed
// outcome. A malformed ref is 400. A missing row and a missing file are both
// 404. A present png or jpeg is 200 with that media type and the immutable
// cache header.
func AttachmentRouteResponseFor(outcome AttachmentRouteOutcome) (AttachmentRouteResponse, error) {
	switch outcome {
	case AttachmentRouteOutcomePNG:
		return attachmentRouteSuccess(ImageMediaTypePNG), nil
	case AttachmentRouteOutcomeJPEG:
		return attachmentRouteSuccess(ImageMediaTypeJPEG), nil
	case AttachmentRouteOutcomeMalformed:
		return AttachmentRouteResponse{
			Status: AttachmentRouteStatusMalformed,
			Code:   AttachmentRouteCodeMalformed,
		}, nil
	case AttachmentRouteOutcomeMissingRow, AttachmentRouteOutcomeMissingFile:
		return AttachmentRouteResponse{
			Status: AttachmentRouteStatusNotFound,
			Code:   AttachmentRouteCodeNotFound,
		}, nil
	default:
		return AttachmentRouteResponse{}, fmt.Errorf(
			"attachment route response failed at schema.AttachmentRouteResponseFor while reading outcome %q: the outcome is outside the closed png|jpeg|malformed|missing_row|missing_file set, so no status is declared; pass a member of AllAttachmentRouteOutcomes",
			outcome,
		)
	}
}

func attachmentRouteSuccess(media ImageMediaType) AttachmentRouteResponse {
	return AttachmentRouteResponse{
		Status:       AttachmentRouteStatusOK,
		MediaType:    media,
		CacheControl: AttachmentCacheControl,
	}
}

package openapi

// The parameters of the segmented transcript reads. The Local and Village API
// builders both embed these beside their own id path field, so each bound,
// default and description has one declaration and the two specs cannot differ.
// The span bounds restate schema.TranscriptWindowSpan{Minimum,Maximum,Default};
// a test holds them equal.

// transcriptSegmentsParams are the parameters of the segment directory read.
type transcriptSegmentsParams struct {
	Page     int    `query:"page" description:"One-based directory page; omitted defaults to 1" minimum:"1"`
	Revision string `query:"revision" description:"Revision to read; omitted serves the current revision"`
}

// transcriptSegmentParams are the parameters of the single segment read.
type transcriptSegmentParams struct {
	Index    int    `path:"index" description:"Segment ordinal from the directory read" minimum:"0"`
	Revision string `query:"revision" description:"Revision to read; omitted serves the current revision"`
}

// transcriptWindowParams are the parameters of the window read.
type transcriptWindowParams struct {
	Anchor   int    `query:"anchor" required:"true" description:"Turn identity to centre the window on" minimum:"0"`
	Span     int    `query:"span" description:"Assistant turns collected on each side, 1..10; omitted defaults to 10" default:"10" minimum:"1" maximum:"10"`
	Revision string `query:"revision" description:"Revision to read; omitted serves the current revision"`
}

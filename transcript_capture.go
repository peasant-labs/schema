package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"unicode/utf8"
)

// TranscriptCaptureRefusal classifies why NewTranscriptCookTurns refused a
// served session detail. It is a Go-only classification: it has no wire form.
type TranscriptCaptureRefusal string

const (
	// TranscriptCaptureDocumentInvalid: the input is not exactly one valid
	// UTF-8 JSON value, or it nests deeper than 10,000 levels anywhere.
	TranscriptCaptureDocumentInvalid TranscriptCaptureRefusal = "document-invalid"
	// TranscriptCaptureDocumentNotObject: the root is valid JSON but not an object.
	TranscriptCaptureDocumentNotObject TranscriptCaptureRefusal = "document-not-object"
	// TranscriptCaptureDocumentDuplicateMember: two root members share a decoded name.
	TranscriptCaptureDocumentDuplicateMember TranscriptCaptureRefusal = "document-duplicate-member"
	// TranscriptCaptureDocumentEnvelope: the root has a sessionDetail member.
	TranscriptCaptureDocumentEnvelope TranscriptCaptureRefusal = "document-envelope"
	// TranscriptCaptureTurnsNotArray: turns is present and neither null nor an array.
	TranscriptCaptureTurnsNotArray TranscriptCaptureRefusal = "turns-not-array"
	// TranscriptCaptureTurnUnicodeEscape: a turn holds an unpaired or invalid
	// surrogate \u escape.
	TranscriptCaptureTurnUnicodeEscape TranscriptCaptureRefusal = "turn-unicode-escape"
	// TranscriptCaptureTurnDuplicateMember: an object in a turn repeats a decoded name.
	TranscriptCaptureTurnDuplicateMember TranscriptCaptureRefusal = "turn-duplicate-member"
	// TranscriptCaptureTurnNotObject: a turn is not a JSON object.
	TranscriptCaptureTurnNotObject TranscriptCaptureRefusal = "turn-not-object"
	// TranscriptCaptureTurnTypeMismatch: a turn cannot be decoded as a TurnDetail.
	TranscriptCaptureTurnTypeMismatch TranscriptCaptureRefusal = "turn-type-mismatch"
)

// AllTranscriptCaptureRefusals is the complete capture refusal set.
var AllTranscriptCaptureRefusals = []TranscriptCaptureRefusal{
	TranscriptCaptureDocumentInvalid,
	TranscriptCaptureDocumentNotObject,
	TranscriptCaptureDocumentDuplicateMember,
	TranscriptCaptureDocumentEnvelope,
	TranscriptCaptureTurnsNotArray,
	TranscriptCaptureTurnUnicodeEscape,
	TranscriptCaptureTurnDuplicateMember,
	TranscriptCaptureTurnNotObject,
	TranscriptCaptureTurnTypeMismatch,
}

// IsValid reports whether v is a capture refusal.
func (v TranscriptCaptureRefusal) IsValid() bool {
	return inSet(v, AllTranscriptCaptureRefusals)
}

// TranscriptCaptureError is every refusal NewTranscriptCookTurns returns.
// Turn is the stored position of the refusing turn, or -1 when the refusal is
// about the document. Err carries the full actionable message.
type TranscriptCaptureError struct {
	Refusal TranscriptCaptureRefusal
	Turn    int
	Err     error
}

func (e *TranscriptCaptureError) Error() string { return e.Err.Error() }

// Unwrap returns the underlying message error.
func (e *TranscriptCaptureError) Unwrap() error { return e.Err }

// transcriptCaptureMaxDepth is encoding/json's nesting limit. A document the
// standard library would call invalid must stay invalid here.
const transcriptCaptureMaxDepth = 10000

// transcriptMember locates one top-level member of a captured turn. The name
// span excludes its quotes; both spans are relative to the turn's first byte.
type transcriptMember struct {
	nameStart, nameEnd, valueStart, valueEnd uint32
}

// transcriptTurnMembers is the exact-name member view of one captured turn.
type transcriptTurnMembers struct {
	raw   []byte
	table []transcriptMember
}

func (m transcriptTurnMembers) member(name string) (json.RawMessage, bool) {
	if uint64(len(m.raw)) > math.MaxUint32 {
		// Offsets of a turn this large do not fit the compact table.
		members, err := transcriptRawMembers(m.raw)
		if err != nil {
			return nil, false
		}
		value, ok := members[name]
		return value, ok
	}
	for _, e := range m.table {
		// An escape always spells its decoded text in more bytes, so a name of
		// the looked-up length matches only byte for byte, and only a longer
		// name can be an escaped spelling.
		n := m.raw[e.nameStart:e.nameEnd]
		switch {
		case len(n) < len(name):
			continue
		case len(n) == len(name):
			if string(n) != name {
				continue
			}
		case bytes.IndexByte(n, '\\') < 0 || transcriptDecodeName(m.raw[e.nameStart-1:e.nameEnd+1]) != name:
			continue
		}
		return json.RawMessage(m.raw[e.valueStart:e.valueEnd:e.valueEnd]), true
	}
	return nil, false
}

// transcriptMemberLookup is an exact-name member view: a captured turn's
// compact table or a decoded member map.
type transcriptMemberLookup interface {
	member(name string) (json.RawMessage, bool)
}

type transcriptMemberMap map[string]json.RawMessage

func (m transcriptMemberMap) member(name string) (json.RawMessage, bool) {
	value, ok := m[name]
	return value, ok
}

// transcriptDecodeName decodes one quoted JSON string with encoding/json, so
// escapes and unpaired surrogates (U+FFFD) decode exactly as the decoder does.
func transcriptDecodeName(quoted []byte) string {
	var name string
	if err := json.Unmarshal(quoted, &name); err != nil {
		// The scanner validated the string; keep a distinct, non-colliding key.
		return "\x00invalid:" + string(quoted)
	}
	return name
}

type captureFault uint8

const (
	captureFaultNone captureFault = iota
	captureFaultDuplicate
	captureFaultEscape
)

type captureTurnsKind uint8

const (
	captureTurnsAbsent captureTurnsKind = iota
	captureTurnsNull
	captureTurnsArray
	captureTurnsOther
)

type captureScratchMember struct {
	nameStart, nameEnd, valueStart, valueEnd int
}

type captureTurn struct {
	start, end             int
	object                 bool
	memberStart, memberEnd int
	fault                  captureFault
	faultName              string
	faultAt                int
}

type captureName struct {
	raw     []byte
	decoded string
	escaped bool
}

func (n captureName) equal(o captureName) bool {
	if !n.escaped && !o.escaped {
		return bytes.Equal(n.raw, o.raw)
	}
	return n.key() == o.key()
}

func (n captureName) key() string {
	if n.escaped {
		return n.decoded
	}
	return string(n.raw)
}

// captureNameSetThreshold is the member count above which duplicate detection
// switches from a linear comparison to a set.
const captureNameSetThreshold = 16

// captureScanner validates a session detail and records its turns in one pass
// over the bytes. It reproduces encoding/json's grammar and depth limit, the
// decoder's member-name decoding, and the turn-level escape and duplicate
// rules, in the order the refusal precedence needs.
type captureScanner struct {
	data  []byte
	pos   int
	depth int

	// Turn state. detailed is false once an earlier turn has a fault, because
	// no later turn can decide the verdict; the rest is checked for syntax only.
	inTurn    bool
	detailed  bool
	turnStart int
	fault     captureFault
	faultName string
	faultAt   int
	names     []captureName
	members   []captureScratchMember
	turns     []captureTurn

	rootObject    bool
	rootDuplicate string
	hasDuplicate  bool
	envelope      bool
	turnsKind     captureTurnsKind
	turnsStart    int
	turnsEnd      int
}

var captureStringPlain = func() (t [256]bool) {
	for c := 0x20; c < 256; c++ {
		t[c] = c != '"' && c != '\\'
	}
	return t
}()

func (s *captureScanner) skipSpace() {
	for s.pos < len(s.data) {
		switch s.data[s.pos] {
		case ' ', '\t', '\n', '\r':
			s.pos++
		default:
			return
		}
	}
}

func captureHex(b []byte) (rune, bool) {
	var v rune
	for _, c := range b {
		switch {
		case '0' <= c && c <= '9':
			v = v<<4 | rune(c-'0')
		case 'a' <= c && c <= 'f':
			v = v<<4 | rune(c-'a'+10)
		case 'A' <= c && c <= 'F':
			v = v<<4 | rune(c-'A'+10)
		default:
			return 0, false
		}
	}
	return v, true
}

// scanString consumes one string at s.pos (an opening quote). It reports
// whether the string holds an escape.
func (s *captureScanner) scanString() (escaped, ok bool) {
	data := s.data
	pos := s.pos + 1
	for {
		for pos < len(data) && captureStringPlain[data[pos]] {
			pos++
		}
		if pos >= len(data) {
			return false, false
		}
		c := data[pos]
		if c == '"' {
			s.pos = pos + 1
			return escaped, true
		}
		if c < 0x20 {
			return false, false
		}
		// c is a backslash.
		escaped = true
		pos++
		if pos >= len(data) {
			return false, false
		}
		switch data[pos] {
		case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			pos++
		case 'u':
			if pos+5 > len(data) {
				return false, false
			}
			v, hex := captureHex(data[pos+1 : pos+5])
			if !hex {
				return false, false
			}
			if s.inTurn && s.fault != captureFaultEscape && v >= 0xD800 && v <= 0xDFFF {
				paired := false
				if v <= 0xDBFF && pos+11 <= len(data) && data[pos+5] == '\\' && data[pos+6] == 'u' {
					if low, hex := captureHex(data[pos+7 : pos+11]); hex && low >= 0xDC00 && low <= 0xDFFF {
						paired = true
					}
				}
				if paired {
					pos += 11
					continue
				}
				s.fault = captureFaultEscape
				s.faultAt = pos - 1 - s.turnStart
			}
			pos += 5
		default:
			return false, false
		}
	}
}

func (s *captureScanner) scanNumber() bool {
	data, pos := s.data, s.pos
	if pos < len(data) && data[pos] == '-' {
		pos++
	}
	if pos >= len(data) {
		return false
	}
	switch c := data[pos]; {
	case c == '0':
		pos++
	case '1' <= c && c <= '9':
		pos++
		for pos < len(data) && '0' <= data[pos] && data[pos] <= '9' {
			pos++
		}
	default:
		return false
	}
	if pos < len(data) && data[pos] == '.' {
		pos++
		start := pos
		for pos < len(data) && '0' <= data[pos] && data[pos] <= '9' {
			pos++
		}
		if pos == start {
			return false
		}
	}
	if pos < len(data) && (data[pos] == 'e' || data[pos] == 'E') {
		pos++
		if pos < len(data) && (data[pos] == '+' || data[pos] == '-') {
			pos++
		}
		start := pos
		for pos < len(data) && '0' <= data[pos] && data[pos] <= '9' {
			pos++
		}
		if pos == start {
			return false
		}
	}
	s.pos = pos
	return true
}

func (s *captureScanner) scanLiteral(word string) bool {
	if !bytes.HasPrefix(s.data[s.pos:], []byte(word)) {
		return false
	}
	s.pos += len(word)
	return true
}

// scanValue consumes one value at s.pos, which must not be whitespace.
func (s *captureScanner) scanValue(turnTop bool) bool {
	if s.pos >= len(s.data) {
		return false
	}
	switch c := s.data[s.pos]; {
	case c == '{':
		return s.scanObject(turnTop)
	case c == '[':
		return s.scanArray()
	case c == '"':
		_, ok := s.scanString()
		return ok
	case c == 't':
		return s.scanLiteral("true")
	case c == 'f':
		return s.scanLiteral("false")
	case c == 'n':
		return s.scanLiteral("null")
	case c == '-' || ('0' <= c && c <= '9'):
		return s.scanNumber()
	default:
		return false
	}
}

func (s *captureScanner) enter() bool {
	s.depth++
	return s.depth <= transcriptCaptureMaxDepth
}

func (s *captureScanner) scanArray() bool {
	if !s.enter() {
		return false
	}
	s.pos++
	s.skipSpace()
	if s.pos < len(s.data) && s.data[s.pos] == ']' {
		s.pos++
		s.depth--
		return true
	}
	for {
		s.skipSpace()
		if !s.scanValue(false) {
			return false
		}
		s.skipSpace()
		if s.pos >= len(s.data) {
			return false
		}
		switch s.data[s.pos] {
		case ',':
			s.pos++
		case ']':
			s.pos++
			s.depth--
			return true
		default:
			return false
		}
	}
}

// scanObject consumes an object. Inside a turn it refuses repeated decoded
// names at every depth, and for the turn's own object records its members.
func (s *captureScanner) scanObject(turnTop bool) bool {
	if !s.enter() {
		return false
	}
	s.pos++
	base := len(s.names)
	defer func() { s.names = s.names[:base] }()
	var set map[string]struct{}
	s.skipSpace()
	if s.pos < len(s.data) && s.data[s.pos] == '}' {
		s.pos++
		s.depth--
		return true
	}
	for {
		s.skipSpace()
		if s.pos >= len(s.data) || s.data[s.pos] != '"' {
			return false
		}
		nameStart := s.pos
		escaped, ok := s.scanString()
		if !ok {
			return false
		}
		nameEnd := s.pos
		if s.inTurn && s.fault == captureFaultNone {
			n := captureName{raw: s.data[nameStart+1 : nameEnd-1], escaped: escaped}
			if escaped {
				n.decoded = transcriptDecodeName(s.data[nameStart:nameEnd])
			}
			duplicate := false
			if set != nil {
				_, duplicate = set[n.key()]
			} else {
				for _, prior := range s.names[base:] {
					if prior.equal(n) {
						duplicate = true
						break
					}
				}
			}
			if duplicate {
				s.fault = captureFaultDuplicate
				s.faultName = n.key()
			} else {
				s.names = append(s.names, n)
				if set != nil {
					set[n.key()] = struct{}{}
				} else if len(s.names)-base > captureNameSetThreshold {
					set = make(map[string]struct{}, 2*captureNameSetThreshold)
					for _, prior := range s.names[base:] {
						set[prior.key()] = struct{}{}
					}
				}
			}
		}
		s.skipSpace()
		if s.pos >= len(s.data) || s.data[s.pos] != ':' {
			return false
		}
		s.pos++
		s.skipSpace()
		valueStart := s.pos
		if !s.scanValue(false) {
			return false
		}
		if turnTop && s.detailed {
			s.members = append(s.members, captureScratchMember{nameStart + 1, nameEnd - 1, valueStart, s.pos})
		}
		s.skipSpace()
		if s.pos >= len(s.data) {
			return false
		}
		switch s.data[s.pos] {
		case ',':
			s.pos++
		case '}':
			s.pos++
			s.depth--
			return true
		default:
			return false
		}
	}
}

// scanTurns consumes the turns array and records each element.
func (s *captureScanner) scanTurns() bool {
	if !s.enter() {
		return false
	}
	s.turnsStart = s.pos
	s.pos++
	s.skipSpace()
	if s.pos < len(s.data) && s.data[s.pos] == ']' {
		s.pos++
		s.depth--
		s.turnsEnd = s.pos
		return true
	}
	for {
		s.skipSpace()
		start := s.pos
		s.inTurn = s.detailed
		s.turnStart = start
		s.fault, s.faultName, s.faultAt = captureFaultNone, "", 0
		memberStart := len(s.members)
		object := s.pos < len(s.data) && s.data[s.pos] == '{'
		if !s.scanValue(true) {
			return false
		}
		s.inTurn = false
		if s.detailed {
			s.turns = append(s.turns, captureTurn{
				start: start, end: s.pos, object: object,
				memberStart: memberStart, memberEnd: len(s.members),
				fault: s.fault, faultName: s.faultName, faultAt: s.faultAt,
			})
			if s.fault != captureFaultNone {
				s.detailed = false
			}
		}
		s.skipSpace()
		if s.pos >= len(s.data) {
			return false
		}
		switch s.data[s.pos] {
		case ',':
			s.pos++
		case ']':
			s.pos++
			s.depth--
			s.turnsEnd = s.pos
			return true
		default:
			return false
		}
	}
}

// scanRoot consumes the root object, refusing nothing itself: it records the
// root duplicate, the envelope and the turns member for the caller to rank.
func (s *captureScanner) scanRoot() bool {
	if !s.enter() {
		return false
	}
	s.pos++
	var names []captureName
	var set map[string]struct{}
	s.skipSpace()
	if s.pos < len(s.data) && s.data[s.pos] == '}' {
		s.pos++
		s.depth--
		return true
	}
	for {
		s.skipSpace()
		if s.pos >= len(s.data) || s.data[s.pos] != '"' {
			return false
		}
		nameStart := s.pos
		escaped, ok := s.scanString()
		if !ok {
			return false
		}
		n := captureName{raw: s.data[nameStart+1 : s.pos-1], escaped: escaped}
		if escaped {
			n.decoded = transcriptDecodeName(s.data[nameStart:s.pos])
		}
		duplicate := false
		if set != nil {
			_, duplicate = set[n.key()]
		} else {
			for _, prior := range names {
				if prior.equal(n) {
					duplicate = true
					break
				}
			}
		}
		if duplicate && !s.hasDuplicate {
			s.hasDuplicate, s.rootDuplicate = true, n.key()
		}
		if !duplicate {
			names = append(names, n)
			if set != nil {
				set[n.key()] = struct{}{}
			} else if len(names) > captureNameSetThreshold {
				set = make(map[string]struct{}, 2*captureNameSetThreshold)
				for _, prior := range names {
					set[prior.key()] = struct{}{}
				}
			}
		}
		key := n.key()
		if key == transcriptWireSessionDetail {
			s.envelope = true
		}
		s.skipSpace()
		if s.pos >= len(s.data) || s.data[s.pos] != ':' {
			return false
		}
		s.pos++
		s.skipSpace()
		if key == transcriptWireTurns && !duplicate {
			switch {
			case s.pos < len(s.data) && s.data[s.pos] == '[':
				s.turnsKind = captureTurnsArray
				if !s.scanTurns() {
					return false
				}
			case s.scanLiteral("null"):
				s.turnsKind = captureTurnsNull
			default:
				s.turnsKind = captureTurnsOther
				if !s.scanValue(false) {
					return false
				}
			}
		} else if !s.scanValue(false) {
			return false
		}
		s.skipSpace()
		if s.pos >= len(s.data) {
			return false
		}
		switch s.data[s.pos] {
		case ',':
			s.pos++
		case '}':
			s.pos++
			s.depth--
			return true
		default:
			return false
		}
	}
}

// scanDocument reports whether data is exactly one valid JSON value.
func (s *captureScanner) scanDocument() bool {
	s.detailed = true
	s.skipSpace()
	if s.pos >= len(s.data) {
		return false
	}
	if s.data[s.pos] == '{' {
		s.rootObject = true
		if !s.scanRoot() {
			return false
		}
	} else if !s.scanValue(false) {
		return false
	}
	s.skipSpace()
	return s.pos == len(s.data)
}

// NewTranscriptCookTurns captures one served session detail's turns in stored
// order. It never unwraps an envelope and never pairs two separate decodes.
// Every refusal is a *TranscriptCaptureError.
//
// The call makes one private copy of the turns array. Every returned turn's
// bytes and member view are slices of that copy, so retaining any one turn
// keeps the whole copy alive. The caller's buffer is never aliased.
func NewTranscriptCookTurns(sessionDetail json.RawMessage) ([]TranscriptCookTurn, error) {
	const where = "transcript turn capture failed at schema.NewTranscriptCookTurns: "
	refuse := func(refusal TranscriptCaptureRefusal, turn int, err error) ([]TranscriptCookTurn, error) {
		return nil, &TranscriptCaptureError{Refusal: refusal, Turn: turn, Err: err}
	}
	s := captureScanner{data: sessionDetail}
	if !utf8.Valid(sessionDetail) || !s.scanDocument() {
		return refuse(TranscriptCaptureDocumentInvalid, -1, fmt.Errorf("%ssession detail is not one valid UTF-8 JSON object; no turns can be paired; pass the served session detail object", where))
	}
	if !s.rootObject {
		return refuse(TranscriptCaptureDocumentNotObject, -1, fmt.Errorf("%ssession detail is not one valid JSON object: %w; pass the served session detail object", where, fmt.Errorf("a JSON object is required")))
	}
	if s.hasDuplicate {
		return refuse(TranscriptCaptureDocumentDuplicateMember, -1, fmt.Errorf("%ssession detail is not one valid JSON object: %w; pass the served session detail object", where, fmt.Errorf("duplicate object key %q", s.rootDuplicate)))
	}
	if s.envelope {
		return refuse(TranscriptCaptureDocumentEnvelope, -1, fmt.Errorf("%san envelope was supplied instead of its session detail; no turns can be paired; select sessionDetail before calling", where))
	}
	switch s.turnsKind {
	case captureTurnsAbsent, captureTurnsNull:
		return []TranscriptCookTurn{}, nil
	case captureTurnsOther:
		return refuse(TranscriptCaptureTurnsNotArray, -1, fmt.Errorf("%sturns is not an array; no turns can be paired; pass the served turn array", where))
	}

	turns := make([]TranscriptCookTurn, len(s.turns))
	table := make([]transcriptMember, len(s.members))
	for i, t := range s.turns {
		if uint64(t.end-t.start) > math.MaxUint32 {
			continue
		}
		for j := t.memberStart; j < t.memberEnd; j++ {
			m := s.members[j]
			table[j] = transcriptMember{
				nameStart: uint32(m.nameStart - t.start), nameEnd: uint32(m.nameEnd - t.start),
				valueStart: uint32(m.valueStart - t.start), valueEnd: uint32(m.valueEnd - t.start),
			}
		}
		turns[i].members = table[t.memberStart:t.memberEnd:t.memberEnd]
	}
	for i, t := range s.turns {
		switch t.fault {
		case captureFaultEscape:
			return refuse(TranscriptCaptureTurnUnicodeEscape, i, fmt.Errorf("%sturn %d has invalid or repeated members: %w; repair the stored turn before retrying", where, i,
				fmt.Errorf("unpaired or invalid Unicode surrogate escape near byte %d of the turn; decoding would replace evidence; use a valid surrogate pair", t.faultAt)))
		case captureFaultDuplicate:
			return refuse(TranscriptCaptureTurnDuplicateMember, i, fmt.Errorf("%sturn %d has invalid or repeated members: %w; repair the stored turn before retrying", where, i,
				fmt.Errorf("duplicate object key %q; map decoding would silently discard evidence; remove the duplicate", t.faultName)))
		}
		if !t.object {
			return refuse(TranscriptCaptureTurnNotObject, i, fmt.Errorf("%sturn %d is not one JSON object: %w; repair the stored turn before retrying", where, i, fmt.Errorf("a JSON object is required")))
		}
		if err := decodeCapturedTurn(sessionDetail[t.start:t.end], turns[i].members, &turns[i].detail); err != nil {
			return refuse(TranscriptCaptureTurnTypeMismatch, i, fmt.Errorf("%sturn %d cannot be decoded as a turn: %w; repair the stored member value before retrying", where, i, err))
		}
	}
	if len(turns) == 0 {
		return turns, nil
	}

	base := s.turnsStart
	held := bytes.Clone(sessionDetail[base:s.turnsEnd])
	for i, t := range s.turns {
		start, end := t.start-base, t.end-base
		turns[i].raw = TranscriptServedTurn{raw: held[start:end:end]}
	}
	return turns, nil
}

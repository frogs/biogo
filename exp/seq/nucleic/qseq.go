// Copyright ©2011-2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package nucleic

import (
	"code.google.com/p/biogo/bio"
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/sequtils"
	"code.google.com/p/biogo/feat"
)

// QSeq is a basic nucleic acid sequence with Phred quality scores.
type QSeq struct {
	ID         string
	Desc       string
	Loc        string
	S          []alphabet.QLetter
	Strand     Strand
	Threshold  alphabet.Qphred // Threshold for returning valid letter.
	LowQFilter seq.Filter      // How to represent below threshold letter.
	Stringify  seq.Stringify   // Function allowing user specified string representation.
	Meta       interface{}     // No operation implicitly copies or changes the contents of Meta.
	alphabet   alphabet.Nucleic
	circular   bool
	offset     int
	encoding   alphabet.Encoding
}

// Create a new QSeq with the given id, letter sequence, alphabet and quality encoding.
func NewQSeq(id string, ql []alphabet.QLetter, alpha alphabet.Nucleic, encode alphabet.Encoding) *QSeq {
	return &QSeq{
		ID:         id,
		S:          append([]alphabet.QLetter(nil), ql...),
		alphabet:   alpha,
		encoding:   encode,
		Strand:     1,
		Threshold:  2,
		LowQFilter: func(s seq.Sequence, _ alphabet.Letter) alphabet.Letter { return s.(*QSeq).alphabet.Ambiguous() },
		Stringify:  QStringify,
	}
}

// Interface guarantees:
var (
	_ seq.Polymer  = &QSeq{}
	_ seq.Sequence = &QSeq{}
	_ seq.Scorer   = &QSeq{}
	_ seq.Appender = &QSeq{}
	_ Sequence     = &QSeq{}
	_ Quality      = &QSeq{}
)

// Required to satisfy nucleic.Sequence interface.
func (self *QSeq) Nucleic() {}

// Name returns a pointer to the ID string of the sequence.
func (self *QSeq) Name() *string { return &self.ID }

// Description returns a pointer to the Desc string of the sequence.
func (self *QSeq) Description() *string { return &self.Desc }

// Location returns a pointer to the Loc string of the sequence.
func (self *QSeq) Location() *string { return &self.Loc }

// Raw returns a pointer to the underlying []Qphred slice.
func (self *QSeq) Raw() interface{} { return &self.S }

// Append QLetters to the sequence, the DefaultQphred value is used for quality scores.
func (self *QSeq) AppendLetters(a ...alphabet.Letter) (err error) {
	l := self.Len()
	self.S = append(self.S, make([]alphabet.QLetter, len(a))...)[:l]
	for _, v := range a {
		self.S = append(self.S, alphabet.QLetter{L: v, Q: DefaultQphred})
	}

	return
}

// Append letters with quality scores to the seq.
func (self *QSeq) AppendQLetters(a ...alphabet.QLetter) (err error) {
	self.S = append(self.S, a...)
	return
}

// Return the Alphabet used by the sequence.
func (self *QSeq) Alphabet() alphabet.Alphabet { return self.alphabet }

// Return the letter at position pos.
func (self *QSeq) At(pos seq.Position) alphabet.QLetter {
	if pos.Ind != 0 {
		panic("nucleic: index out of range")
	}
	return self.S[pos.Pos-self.offset]
}

// Encode the quality at position pos to a letter based on the sequence encoding setting.
func (self *QSeq) QEncode(pos seq.Position) byte {
	if pos.Ind != 0 {
		panic("nucleic: index out of range")
	}
	return self.S[pos.Pos-self.offset].Q.Encode(self.encoding)
}

// Decode a quality letter to a phred score based on the sequence encoding setting.
func (self *QSeq) QDecode(l byte) alphabet.Qphred { return alphabet.DecodeToQphred(l, self.encoding) }

// Return the quality encoding type.
func (self *QSeq) Encoding() alphabet.Encoding { return self.encoding }

// Set the quality encoding type to e.
func (self *QSeq) SetEncoding(e alphabet.Encoding) { self.encoding = e }

// Return the probability of a sequence error at position pos.
func (self *QSeq) EAt(pos seq.Position) float64 {
	if pos.Ind != 0 {
		panic("nucleic: index out of range")
	}
	return self.S[pos.Pos-self.offset].Q.ProbE()
}

// Set the letter at position pos to l.
func (self *QSeq) Set(pos seq.Position, l alphabet.QLetter) {
	if pos.Ind != 0 {
		panic("nucleic: index out of range")
	}
	self.S[pos.Pos-self.offset] = l
}

// Set the quality at position pos to e to reflect the given p(Error).
func (self *QSeq) SetE(pos seq.Position, e float64) {
	if pos.Ind != 0 {
		panic("nucleic: index out of range")
	}
	self.S[pos.Pos-self.offset].Q = alphabet.Ephred(e)
}

// Return the length of the sequence.
func (self *QSeq) Len() int { return len(self.S) }

// Satisfy Counter.
func (self *QSeq) Count() int { return 1 }

// Set the global offset of the sequence to o.
func (self *QSeq) Offset(o int) { self.offset = o }

// Return the start position of the sequence in global coordinates.
func (self *QSeq) Start() int { return self.offset }

// Return the end position of the sequence in global coordinates.
func (self *QSeq) End() int { return self.offset + self.Len() }

// Return the molecule type of the sequence.
func (self *QSeq) Moltype() bio.Moltype { return self.alphabet.Moltype() }

// Validate the letters of the sequence according to the specified alphabet.
func (self *QSeq) Validate() (bool, int) {
	for i, ql := range self.S {
		if !self.alphabet.IsValid(ql.L) {
			return false, i
		}
	}

	return true, -1
}

// Return a copy of the sequence.
func (self *QSeq) Copy() seq.Sequence {
	c := *self
	c.S = append([]alphabet.QLetter(nil), self.S...)
	c.Meta = nil

	return &c
}

// Reverse complement the sequence.
func (self *QSeq) RevComp() {
	self.S = self.revComp(self.S, self.alphabet.ComplementTable())
	self.Strand = -self.Strand
}

func (self *QSeq) revComp(s []alphabet.QLetter, complement []alphabet.Letter) []alphabet.QLetter {
	i, j := 0, len(s)-1
	for ; i < j; i, j = i+1, j-1 {
		s[i].L, s[j].L = complement[s[j].L], complement[s[i].L]
		s[i].Q, s[j].Q = s[j].Q, s[i].Q
	}
	if i == j {
		s[i].L = complement[s[i].L]
	}

	return s
}

// Reverse the sequence.
func (self *QSeq) Reverse() { self.S = sequtils.Reverse(self.S).([]alphabet.QLetter) }

// Specify that the sequence is circular.
func (self *QSeq) Circular(c bool) { self.circular = c }

// Return whether the sequence is circular.
func (self *QSeq) IsCircular() bool { return self.circular }

// Return a subsequence from start to end, wrapping if the sequence is circular.
func (self *QSeq) Subseq(start int, end int) (sub seq.Sequence, err error) {
	var s *QSeq

	tt, err := sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular)
	if err == nil {
		s = &QSeq{}
		*s = *self
		s.S = tt.([]alphabet.QLetter)
		s.S = nil
		s.Meta = nil
		s.offset = start
		s.circular = false
	}

	return s, nil
}

// Truncate the sequenc from start to end, wrapping if the sequence is circular.
func (self *QSeq) Truncate(start int, end int) (err error) {
	tt, err := sequtils.Truncate(self.S, start-self.offset, end-self.offset, self.circular)
	if err == nil {
		self.S = tt.([]alphabet.QLetter)
		self.offset = start
		self.circular = false
	}

	return
}

// Join p to the sequence at the end specified by where.
func (self *QSeq) Join(p *QSeq, where int) (err error) {
	if self.circular {
		return bio.NewError("Cannot join circular sequence: receiver.", 1, self)
	} else if p.circular {
		return bio.NewError("Cannot join circular sequence: parameter.", 1, p)
	}

	var tt interface{}

	tt, self.offset = sequtils.Join(self.S, p.S, where)
	self.S = tt.([]alphabet.QLetter)

	return
}

// Join sequentially order disjunct segments of the sequence, returning any error.
func (self *QSeq) Stitch(f feat.FeatureSet) (err error) {
	tt, err := sequtils.Stitch(self.S, self.offset, f)
	if err == nil {
		self.S = tt.([]alphabet.QLetter)
		self.circular = false
		self.offset = 0
	}

	return
}

// Join segments of the sequence, returning any error.
func (self *QSeq) Compose(f feat.FeatureSet) (err error) {
	tt, err := sequtils.Compose(self.S, self.offset, f)
	if err == nil {
		s := []alphabet.QLetter{}
		complement := self.alphabet.ComplementTable()
		for i, ts := range tt {
			if f[i].Strand == -1 {
				s = append(s, self.revComp(ts.([]alphabet.QLetter), complement)...)
			} else {
				s = append(s, ts.([]alphabet.QLetter)...)
			}
		}

		self.S = s
		self.circular = false
		self.offset = 0
	}

	return
}

// Return a string representation of the sequence. Representation is determined by the Stringify field.
func (self *QSeq) String() string { return self.Stringify(self) }

// The default Stringify function for QSeq.
var QStringify = func(s seq.Polymer) string {
	t := s.(*QSeq)
	gap := t.Alphabet().Gap()
	cs := make([]alphabet.Letter, 0, len(t.S))
	for _, ql := range t.S {
		if alphabet.Qphred(ql.Q) > t.Threshold || ql.L == gap {
			cs = append(cs, ql.L)
		} else {
			cs = append(cs, t.LowQFilter(t, ql.L))
		}
	}

	return alphabet.Letters(cs).String()
}

// The default LowQFilter function for QSeq.
var LowQFilter = func(s seq.Sequence, _ alphabet.Letter) alphabet.Letter { return s.(*QSeq).alphabet.Ambiguous() }

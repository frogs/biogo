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

package pals

import (
	"code.google.com/p/biogo/feat"
	"code.google.com/p/biogo/io/featio/gff"
	"fmt"
	"io"
	"os"
)

var t *feat.Feature = &feat.Feature{Source: "pals", Feature: "hit"}

// PALS pair writer type.
type Writer struct {
	w *gff.Writer
}

// Returns a new PALS writer using f.
func NewWriter(f io.WriteCloser, v, width int, header bool) (w *Writer) {
	return &Writer{gff.NewWriter(f, v, width, header)}
}

// Returns a new PALS writer using a filename, truncating any existing file.
// If appending is required use NewWriter and os.OpenFile.
func NewWriterName(name string, v, width int, header bool) (*Writer, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	return NewWriter(f, v, width, header), nil
}

// Write a single feature and return the number of bytes written and any error.
func (w *Writer) Write(pair *FeaturePair) (n int, err error) {
	t.Location = pair.B.ID
	t.Start = pair.B.Start
	t.End = pair.B.End
	t.Score = floatPtr(float64(pair.Score))
	t.Strand = pair.Strand
	t.Frame = -1
	t.Attributes = fmt.Sprintf("Target %s %d %d; maxe %.2g", pair.A.ID, pair.A.Start+1, pair.A.End, pair.Error) // +1 is kludge for absence of gffwriter
	return w.w.Write(t)
}

func floatPtr(f float64) *float64 { return &f }

// Close the writer, flushing any unwritten data.
func (w *Writer) Close() (err error) {
	return w.w.Close()
}

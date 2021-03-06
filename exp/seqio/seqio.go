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

// Packages for reading and writing sequence files
package seqio

import "code.google.com/p/biogo/exp/seq"

// A Sequence is a generic sequence type.
type Sequence interface {
	seq.Polymer
	seq.Sequence
}

// A SequenceAppender is a generic sequence type that can append elements.
type SequenceAppender interface {
	seq.Appender
	Sequence
}

type Reader interface {
	Read() (seq.Sequence, error)
}

type Writer interface {
	Write(Sequence) (int, error)
}

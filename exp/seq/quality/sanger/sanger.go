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

//Package sanger provides support for manipulation of quality data in Sanger format.
package sanger

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/quality"
)

type Sanger struct {
	*quality.Phred
}

func NewSanger(id string, q []alphabet.Qphred) *Sanger {
	return &Sanger{quality.NewPhred(id, q, alphabet.Sanger)}
}

func (self *Sanger) Join(p *Sanger, where int) (err error) {
	return self.Phred.Join(p.Phred, where)
}

func (self *Sanger) Copy() seq.Quality {
	return &Sanger{self.Phred.Copy().(*quality.Phred)}
}

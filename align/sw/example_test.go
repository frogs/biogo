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

package sw

import (
	"code.google.com/p/biogo/seq"
	"fmt"
)

func ExampleAligner_Align() {
	swsa := &seq.Seq{Seq: []byte("ACACACTA")}
	swsb := &seq.Seq{Seq: []byte("AGCACACA")}

	// w(gap) = -1
	// w(match) = +2
	// w(mismatch) = -1
	swm := [][]int{
		{2, -1, -1, -1, -1},
		{-1, 2, -1, -1, -1},
		{-1, -1, 2, -1, -1},
		{-1, -1, -1, 2, -1},
		{-1, -1, -1, -1, 0},
	}

	smith := &Aligner{Matrix: swm, LookUp: LookUpN, GapChar: '-'}
	swa, err := smith.Align(swsa, swsb)
	if err == nil {
		fmt.Printf("%s\n%s\n", swa[0].Seq, swa[1].Seq)
	}
	// Output:
	// A-CACACTA
	// AGCACAC-A
}

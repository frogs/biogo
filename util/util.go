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

// Commonly used functions
package util

// A CTL hold a code to letter lookup table.
// TODO: Replace this with the functionality provided by alphabet.
type CTL struct {
	ValueToCode [256]int
}

// Inititialise and return a CTL based on a map m.
func NewCTL(m map[int]int) (t *CTL) {
	t = &CTL{}
	for i := 0; i < 256; i++ {
		if code, present := m[i]; present {
			t.ValueToCode[i] = code
		} else {
			t.ValueToCode[i] = -1
		}
	}

	return
}

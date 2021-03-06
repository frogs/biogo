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

package multi

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq"
	"code.google.com/p/biogo/exp/seq/protein"
	"code.google.com/p/biogo/feat"
	"fmt"
	"strings"
)

var (
	m, n    *Multi
	aligned = func(a *Multi) {
		start := a.Start()
		for i := 0; i < a.Count(); i++ {
			s := a.Get(i)
			fmt.Printf("%s%v\n", strings.Repeat(" ", s.Start()-start), s)
		}
		fmt.Println()
		fmt.Println(a)
	}
)

func init() {
	var err error
	m, err = NewMulti("example multi",
		[]protein.Sequence{
			protein.NewSeq("example protein 1", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.Protein),
			protein.NewSeq("example protein 2", []alphabet.Letter("ACGGTGACCTGGCGCGCAT"), alphabet.Protein),
			protein.NewSeq("example protein 3", []alphabet.Letter("ACGATGACGTGGCGCTCAT"), alphabet.Protein),
		},
		protein.Consensify)

	if err != nil {
		panic(err)
	}
}

func ExampleNewMulti() {
	m, err := NewMulti("example multi",
		[]protein.Sequence{
			protein.NewSeq("example protein 1", []alphabet.Letter("ACGCTGACTTGGTGCACGT"), alphabet.Protein),
			protein.NewSeq("example protein 2", []alphabet.Letter("ACGGTGACCTGGCGCGCAT"), alphabet.Protein),
			protein.NewSeq("example protein 3", []alphabet.Letter("ACGATGACGTGGCGCTCAT"), alphabet.Protein),
		},
		protein.Consensify)

	if err != nil {
		return
	}

	aligned(m)
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// 
	// acgxtgacxtggcgcxcat
}

func ExampleMulti_Add() {
	fmt.Printf("%v %v\n", m.Count(), m)
	m.Add(protein.NewQSeq("example protein",
		[]alphabet.QLetter{{'a', 40}, {'c', 39}, {'g', 40}, {'C', 38}, {'t', 35}, {'g', 20}},
		alphabet.Protein, alphabet.Sanger))
	fmt.Printf("%v %v\n", m.Count(), m)
	// Output:
	// 3 acgxtgacxtggcgcxcat
	// 4 acgctgacxtggcgcxcat
}

func ExampleMulti_Copy() {
	n = m.Copy().(*Multi)
	n.Set(seq.Position{Pos: 3, Ind: 2}, alphabet.QLetter{L: 't'})
	aligned(m)
	fmt.Println()
	aligned(n)
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// acgCtg
	// 
	// acgctgacxtggcgcxcat
	// 
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCAT
	// acgCtg
	// 
	// acgctgacxtggcgcxcat
}

func ExampleMulti_Count() {
	fmt.Println(m.Count())
	// Output:
	// 4
}

func ExampleMulti_IsFlush() {
	m.Get(3).Offset(13)
	aligned(m)
	fmt.Printf("\nFlush at left: %v\nFlush at right: %v\n", m.IsFlush(seq.Start), m.IsFlush(seq.End))
	m.Flush(seq.Start, '-')
	fmt.Println()
	aligned(m)
	fmt.Printf("\nFlush at left: %v\nFlush at right: %v\n", m.IsFlush(seq.Start), m.IsFlush(seq.End))
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	//              acgCtg
	// 
	// acgxtgacxtggcgcgcat
	// 
	// Flush at left: false
	// Flush at right: true
	// 
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// -------------acgCtg
	// 
	// acgxtgacxtggcgcgcat
	//
	// Flush at left: true
	// Flush at right: true
}

func ExampleMulti_Join() {
	aligned(n)
	n.Join(m, seq.End)
	fmt.Println()
	aligned(n)
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCAT
	// acgCtg
	// 
	// acgctgacxtggcgcxcat
	// 
	// ACGCTGACTTGGTGCACGTACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCATACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCATACGATGACGTGGCGCTCAT
	// acgCtg--------------------------acgCtg
	// 
	// acgctgacxtggcgcxcatacgxtgacxtggcgcgcat
}

func ExampleMulti_Len() {
	fmt.Println(m.Len())
	// Output:
	// 19
}

func ExampleMulti_Reverse() {
	aligned(m)
	fmt.Println()
	m.Reverse()
	aligned(m)
	// Output:
	// ACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCAT
	// ACGATGACGTGGCGCTCAT
	// -------------acgCtg
	// 
	// acgxtgacxtggcgcgcat
	// 
	// TGCACGTGGTTCAGTCGCA
	// TACGCGCGGTCCAGTGGCA
	// TACTCGCGGTGCAGTAGCA
	// gtCgca-------------
	// 
	// tacgcgcggtxcagtxgca
}

func ExampleMulti_Stitch() {
	f := feat.FeatureSet{
		&feat.Feature{Start: -1, End: 4},
		&feat.Feature{Start: 30, End: 38},
	}
	aligned(n)
	fmt.Println()
	err := n.Stitch(f)
	if err != nil {
		fmt.Println(err)
	} else {
		aligned(n)
	}
	// Output:
	// ACGCTGACTTGGTGCACGTACGCTGACTTGGTGCACGT
	// ACGGTGACCTGGCGCGCATACGGTGACCTGGCGCGCAT
	// ACGtTGACGTGGCGCTCATACGATGACGTGGCGCTCAT
	// acgCtg--------------------------acgCtg
	// 
	// acgctgacxtggcgcxcatacgxtgacxtggcgcgcat
	//
	// ACGCGTGCACGT
	// ACGGGCGCGCAT
	// ACGtGCGCTCAT
	// acgC--acgCtg
	// 
	// acgcgcgcgcat
}

func ExampleMulti_Truncate() {
	aligned(m)
	m.Truncate(4, 12)
	fmt.Println()
	aligned(m)
	// Output:
	// TGCACGTGGTTCAGTCGCA
	// TACGCGCGGTCCAGTGGCA
	// TACTCGCGGTGCAGTAGCA
	// gtCgca-------------
	// 
	// tacgcgcggtxcagtxgca
	// 
	// CGTGGTTC
	// CGCGGTCC
	// CGCGGTGC
	// ca------
	// 
	// cgcggtxc
}

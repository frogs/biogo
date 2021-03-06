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

// Package implementing functions required for PALS sequence alignment
package pals

import (
	"code.google.com/p/biogo/align/pals/dp"
	"code.google.com/p/biogo/align/pals/filter"
	"code.google.com/p/biogo/bio"
	"code.google.com/p/biogo/index/kmerindex"
	"code.google.com/p/biogo/morass"
	"code.google.com/p/biogo/seq"
	"code.google.com/p/biogo/util"
	"io"
	"os"
	"unsafe"
)

// Default values for filter and alignment.
var (
	MaxIGap    = 5
	DiffCost   = 3
	SameCost   = 1
	MatchCost  = DiffCost + SameCost
	BlockCost  = DiffCost * MaxIGap
	RMatchCost = float64(DiffCost) + 1
)

// Default thresholds for filter and alignment.
var (
	DefaultLength      = 400
	DefaultMinIdentity = 0.94
	MaxAvgIndexListLen = 15.0
	TubeOffsetDelta    = 32
)

// Default word characteristics.
var (
	MinWordLength = 4  // For minimum word length, choose k=4 arbitrarily.
	MaxKmerLen    = 15 // Currently limited to 15 due to 32 bit int limit for indexing slices
)

// PALS is a type that can perform pairwise alignments of large sequences based on the papers:
//  PILER: identification and classification of genomic repeats.
//   Robert C. Edgar and Eugene W. Myers. Bioinformatics Suppl. 1:i152-i158 (2005)
//  Efficient q-gram filters for finding all 𝛜-matches over a given length.
//   Kim R. Rasmussen, Jens Stoye, and Eugene W. Myers. J. of Computational Biology 13:296–308 (2006).
type PALS struct {
	target, query *seq.Seq
	selfCompare   bool
	index         *kmerindex.Index
	FilterParams  *filter.Params
	DPParams      *dp.Params
	MaxIGap       int
	DiffCost      int
	SameCost      int
	MatchCost     int
	BlockCost     int
	RMatchCost    float64

	log        Logger
	timer      *util.Timer
	tubeOffset int
	maxMem     *uintptr
	hitFilter  *filter.Filter
	morass     *morass.Morass
	err        error
	threads    int
}

// Return a new PALS aligner. Requires
func New(target, query *seq.Seq, selfComp bool, m *morass.Morass, threads, tubeOffset int, mem *uintptr, log Logger) *PALS {
	return &PALS{
		target:      target,
		query:       query,
		selfCompare: selfComp,
		log:         log,
		tubeOffset:  tubeOffset,
		MaxIGap:     MaxIGap,
		DiffCost:    DiffCost,
		SameCost:    SameCost,
		MatchCost:   MatchCost,
		BlockCost:   BlockCost,
		RMatchCost:  RMatchCost,
		maxMem:      mem,
		morass:      m,
		threads:     threads,
	}
}

// Optimise the PALS parameters for given memory, kmer length, hit length and sequence identity.
// An error is returned if no satisfactory parameters can be found.
func (p *PALS) Optimise(minHitLen int, minId float64) error {
	if minId < 0 || minId > 1.0 {
		return bio.NewError("bad minId", 0, minId)
	}
	if minHitLen <= MinWordLength {
		return bio.NewError("bad minHitLength", 0, minHitLen)
	}

	if p.log != nil {
		p.log.Print("Optimising filter parameters")
	}

	filterParams := &filter.Params{}

	// Lower bound on word length k by requiring manageable index.
	// Given kmer occurs once every 4^k positions.
	// Hence average number of index entries is i = N/(4^k) for random
	// string of length N.
	// Require i <= I, then k > log_4(N/i).
	minWordSize := int(util.Log4(float64(p.target.Len())) - util.Log4(MaxAvgIndexListLen) + 0.5)

	// First choice is that filter criteria are same as DP criteria,
	// but this may not be possible.
	seedLength := minHitLen
	seedDiffs := int(float64(minHitLen) * (1 - minId))

	// Find filter valid filter parameters, starting from preferred case.
	for {
		minWords := -1
		if MaxKmerLen < minWordSize {
			if p.log != nil {
				p.log.Printf("Word size too small: %d < %d\n", MaxKmerLen, minWordSize)
			}
		}
		for wordSize := MaxKmerLen; wordSize >= minWordSize; wordSize-- {
			filterParams.WordSize = wordSize
			filterParams.MinMatch = seedLength
			filterParams.MaxError = seedDiffs
			if p.tubeOffset > 0 {
				filterParams.TubeOffset = p.tubeOffset
			} else {
				filterParams.TubeOffset = filterParams.MaxError + TubeOffsetDelta
			}

			mem := p.MemRequired(filterParams)
			if p.maxMem != nil && mem > *p.maxMem {
				if p.log != nil {
					p.log.Printf("Parameters n=%d k=%d e=%d, mem=%d MB > maxmem=%d MB\n",
						filterParams.MinMatch,
						filterParams.WordSize,
						filterParams.MaxError,
						mem/1e6,
						*p.maxMem/1e6)
				}
				minWords = -1
				continue
			}

			minWords = filter.MinWordsPerFilterHit(seedLength, wordSize, seedDiffs)
			if minWords <= 0 {
				if p.log != nil {
					p.log.Printf("Parameters n=%d k=%d e=%d, B=%d\n",
						filterParams.MinMatch,
						filterParams.WordSize,
						filterParams.MaxError,
						minWords)
				}
				minWords = -1
				continue
			}

			length := p.AvgIndexListLength(filterParams)
			if length > MaxAvgIndexListLen {
				if p.log != nil {
					p.log.Printf("Parameters n=%d k=%d e=%d, B=%d avgixlen=%.2f > max = %.2f\n",
						filterParams.MinMatch,
						filterParams.WordSize,
						filterParams.MaxError,
						minWords,
						length,
						MaxAvgIndexListLen)
				}
				minWords = -1
				continue
			}
			break
		}
		if minWords > 0 {
			break
		}

		// Failed to find filter parameters, try
		// fewer errors and shorter seed.
		if seedLength >= minHitLen/4 {
			seedLength /= 2
			continue
		}
		if seedDiffs > 0 {
			seedDiffs--
			continue
		}

		return bio.NewError("failed to find filter parameters", 0)
	}

	p.FilterParams = filterParams

	p.DPParams = &dp.Params{
		MinHitLength: minHitLen,
		MinId:        minId,
	}

	return nil
}

// Return an estimate of the average number of hits for any given kmer.
func (p *PALS) AvgIndexListLength(filterParams *filter.Params) float64 {
	return float64(p.target.Len()) / float64(int(1)<<(uint(filterParams.WordSize)*2))
}

// Return an estimate of the amount of memory required for the filter.
func (p *PALS) filterMemRequired(filterParams *filter.Params) uintptr {
	words := util.Pow4(filterParams.WordSize)
	tubeWidth := filterParams.TubeOffset + filterParams.MaxError
	maxActiveTubes := (p.target.Len()+tubeWidth-1)/filterParams.TubeOffset + 1
	tubes := uintptr(maxActiveTubes) * unsafe.Sizeof(tubeState{})
	finger := unsafe.Sizeof(uint32(0)) * uintptr(words)
	pos := unsafe.Sizeof(0) * uintptr(p.target.Len())

	return finger + pos + tubes
}

// filter.tubeState is repeated here to allow memory calculation without exporting tubeState from filter package.
type tubeState struct {
	QLo   int
	QHi   int
	Count int
}

// Return an estimate of the total amount of memory required.
func (p *PALS) MemRequired(filterParams *filter.Params) uintptr {
	filter := p.filterMemRequired(filterParams)
	sequence := uintptr(p.target.Len()) + unsafe.Sizeof(p.target)
	if p.target != p.query {
		sequence += uintptr(p.query.Len()) + unsafe.Sizeof(p.query)
	}

	return filter + sequence
}

// Build the kmerindex for filtering.
func (p *PALS) BuildIndex() error {
	p.notify("Indexing")
	index, err := kmerindex.New(p.FilterParams.WordSize, p.target)
	if err != nil {
		return err
	} else {
		index.Build()
		p.notify("Indexed")
	}
	p.index = index
	p.hitFilter = filter.New(p.index, p.FilterParams)

	return nil
}

// Share allows the receiver to use the index and parameters of m.
func (p *PALS) Share(m *PALS) {
	p.index = m.index
	p.FilterParams = m.FilterParams
	p.DPParams = m.DPParams
	p.hitFilter = filter.New(p.index, p.FilterParams)
}

// Perform filtering and alignment for one strand of query.
func (p *PALS) Align(complement bool) (dp.DPHits, error) {
	if p.err != nil {
		return nil, p.err
	}
	var (
		working *seq.Seq
		err     error
	)
	if complement {
		p.notify("Complementing query")
		working, _ = p.query.RevComp()
		p.notify("Complemented query")
	} else {
		working = p.query
	}

	p.notify("Filtering")
	err = p.hitFilter.Filter(working, p.selfCompare, complement, p.morass)
	if err != nil {
		return nil, err
	}
	p.notifyf("Identified %d filter hits", p.morass.Len())

	p.notify("Merging")
	merger := filter.NewMerger(p.index, working, p.FilterParams, p.MaxIGap, p.selfCompare)
	var hit filter.FilterHit
	for {
		if err = p.morass.Pull(&hit); err != nil {
			break
		}
		merger.MergeFilterHit(&hit)
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	p.err = p.morass.Clear()
	trapezoids := merger.FinaliseMerge()
	lt, lq := trapezoids.Sum()
	p.notifyf("Merged %d trapezoids covering %d x %d", len(trapezoids), lt, lq)

	p.notify("Aligning")
	aligner := dp.NewAligner(p.target, working, p.FilterParams.WordSize, p.DPParams.MinHitLength, p.DPParams.MinId)
	aligner.Config = &dp.AlignConfig{
		MaxIGap:    p.MaxIGap,
		DiffCost:   p.DiffCost,
		SameCost:   p.SameCost,
		MatchCost:  p.MatchCost,
		BlockCost:  p.BlockCost,
		RMatchCost: p.RMatchCost,
	}
	hits := aligner.AlignTraps(trapezoids)
	hitCoverageA, hitCoverageB, err := hits.Sum()
	if err != nil {
		return nil, err
	}
	p.notifyf("Aligned %d hits covering %d x %d", len(hits), hitCoverageA, hitCoverageB)

	return hits, nil
}

// Remove filesystem components of filter. This should be called after the last use of the aligner.
func (p *PALS) CleanUp() error { return p.morass.CleanUp() }

// Interface for logger used by PALS.
type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
}

func (p *PALS) notify(n string) {
	if p.log != nil {
		p.log.Print(n)
	}
}

func (p *PALS) notifyf(f string, n ...interface{}) {
	if p.log != nil {
		p.log.Printf(f, n...)
	}
}

func (p *PALS) fatal(n string) {
	if p.log != nil {
		p.log.Fatal(n)
	}
	os.Exit(1)
}

func (p *PALS) fatalf(f string, n ...interface{}) {
	if p.log != nil {
		p.log.Fatalf(f, n...)
	}
	os.Exit(1)
}

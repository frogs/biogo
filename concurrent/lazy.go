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

package concurrent

// Evaluator is a function for lazy evaluation.
type Evaluator func(...interface{}) (interface{}, State)

type State []interface{}

// Lazily is function to generate a lazy evaluator.
// 
// Lazy functions are terminated by closing the reaper channel. nil should be passed as
// a reaper for perpetual lazy functions.
func Lazily(f Evaluator, lookahead int, reaper <-chan struct{}, init ...interface{}) func() interface{} {
	rc := make(chan interface{}, lookahead)
	go func(rc chan interface{}) {
		defer close(rc)
		var state State = init
		var result interface{}

		for {
			result, state = f(state...)
			select {
			case rc <- result:
			case <-reaper:
				return
			}
		}
	}(rc)

	return func() interface{} {
		return <-rc
	}
}

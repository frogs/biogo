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

package util

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	check "launchpad.net/gocheck"
	"os"
	"os/exec"
)

// Tests
func (s *S) TestHash(c *check.C) {
	// FIXME: This will not work with MacOS.
	_, err := exec.LookPath("md5sum")
	if err != nil {
		c.Skip(err.Error())
	}
	md5sum := exec.Command("md5sum", "./files_test.go")
	b := &bytes.Buffer{}
	md5sum.Stdout = b
	err = md5sum.Run()
	if err != nil {
		c.Fatal(err)
	}
	f, err := os.Open("./files_test.go")
	if err != nil {
		c.Fatalf("%v %s", md5sum, err)
	}
	x, err := ioutil.ReadAll(f)
	if err != nil {
		c.Fatal(err)
	}
	f.Seek(0, 0)

	md5hash, err := Hash(md5.New(), f)
	if err != nil {
		c.Fatal(err)
	}
	md5string := fmt.Sprintf("%x .*\n", md5hash)

	c.Check(string(b.Bytes()), check.Matches, md5string)

	y, err := ioutil.ReadAll(f)
	if err != nil {
		c.Fatal(err)
	}
	c.Check(x, check.DeepEquals, y)
}

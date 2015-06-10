// Copyright (C) 2015 Vinay Kumar
//
// zipscan is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// zipscan is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

package main

import (
	"bufio"
	"io"
	"regexp"
)

type stringFinder func(string) bool
type contentFinder func(io.ReadCloser) bool

func NewStringFinder(pPattern *regexp.Regexp) stringFinder {
	return func(pSourceString string) bool {

		if pPattern == nil {
			return false
		}

		match := pPattern.FindStringIndex(pSourceString)
		if match == nil {
			return false
		}
		return true
	}
}

func NewContentFinder(pPattern *regexp.Regexp) contentFinder {
	return func(pReader io.ReadCloser) bool {
		defer pReader.Close()

		if pPattern == nil {
			return false
		}

		bReader := bufio.NewReader(pReader)
		match := pPattern.FindReaderIndex(bReader)
		if match == nil {
			return false
		}
		return true
	}
}

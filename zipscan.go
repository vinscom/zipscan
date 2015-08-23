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
	"runtime"
	"errors"
	"flag"
	"fmt"
	"regexp"
	"strings"
)

type fileInfoChannel chan fileInfo
type csvString []string

func (i *csvString) String() string {
	return fmt.Sprint(*i)
}

func (i *csvString) Set(value string) error {
	if len(*i) > 0 {
		return errors.New("flag already set")
	}
	for _, s := range strings.Split(value, ",") {
		*i = append(*i, s)
	}
	return nil
}

func main() {

	//Parse Arguments
	var argFileFilterPattern csvString

	flag.Var(&argFileFilterPattern, "f", "Comma seperated list of patterns of name. Only matching names will be search for content or name")
	argTargetDirectory := flag.String("d", ".", "Directory to scan (Symbolic Links are not followed)")
	argSearchContent := flag.Bool("s", false, "Enable content search. To eanble add -s=true or -s")
	argNamePattern := flag.String("p", ".*", "Regular expression of name of file or directory")
	argContentPattern := flag.String("c", "^$", "Regular expression we are looking in file ")

	flag.Parse()
	//Parse Arguments

	//Setup Environment Config
	patternContent, _ := regexp.Compile(*argContentPattern)

	//If content search is enabled then content pattern is also file pattern
	patternName, _ := regexp.Compile(*argNamePattern)
	
	//Setup Environment Config

	chain := make([]processorInfo, 0)

	fnFileNameMatcher := NewStringFinder(patternName)
	fnFileContentMatcher := NewContentFinder(patternContent)

	chain = append(chain,
		processorInfo{NewProcessor(NewTraverseDirectoryProcessor(*argTargetDirectory)), 1},
		processorInfo{NewProcessor(NewFileNameFilterProcessor(argFileFilterPattern)), 1},
		processorInfo{NewProcessor(NewZipFileProcessor(fnFileContentMatcher, *argSearchContent)), runtime.NumCPU()},
		processorInfo{NewProcessor(NewNormalFileProcessor(fnFileContentMatcher, *argSearchContent)), runtime.NumCPU()},
		processorInfo{NewProcessor(NewMatchFileNameProcessor(fnFileNameMatcher, !(*argSearchContent))), 1},
		processorInfo{NewProcessor(PrintToConsoleProcessor), 1})

	done := SetupSystem(&chain)

	<-done
}

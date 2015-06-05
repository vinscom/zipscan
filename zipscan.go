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
	"archive/zip"
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

type fileInfo struct {
	path              string
	dir               bool
	zip               bool
	foundContentMatch bool
	foundPathMatch    bool
}

type stringFinder func(string) bool
type contentFinder func(io.Reader) bool
type fileScanner func(string) []fileInfo
type filterFile func(string) bool
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
	argSearchContent := flag.Bool("s", false, "Enable content search. If this is enabled then Content and File Name patterns become same")
	argNamePattern := flag.String("p", ".*", "Regular expression of name of file or directory")
	argContentPattern := flag.String("c", "^$", "Regular expression we are looking in file ")

	flag.Parse()
	//Parse Arguments

	//Setup Environment Config
	runtime.GOMAXPROCS(2)
	chanFileInfo := make(chan fileInfo, 1)
	chanDone := make(chan bool)
	patternContent, _ := regexp.Compile(*argContentPattern)
	
	//If content search is enabled then content pattern is also file pattern
	var patternName *regexp.Regexp
	
	if *argSearchContent {
		patternName = patternContent	
	} else {
		patternName, _ = regexp.Compile(*argNamePattern)	
	}
	//Setup Environment Config

	go PrintFileInfoChannel(chanFileInfo, chanDone)

	filepath.Walk(*argTargetDirectory, NewWalker(patternName, patternContent, chanFileInfo, argFileFilterPattern, *argSearchContent))

	close(chanFileInfo)

	<-chanDone
}

func PrintFileInfoChannel(pInFileInfoChannel chan fileInfo, pDone chan bool) {
	for {
		select {
		case fInfo, ok := <-pInFileInfoChannel:
			if ok {
				if fInfo.foundContentMatch || fInfo.foundPathMatch {
					fmt.Println(fInfo.path)
				}
			} else {
				pDone <- true
				break
			}
		}
	}
}

func NewWalker(pNamePattern *regexp.Regexp, pContentPattern *regexp.Regexp, pOutFileInfoChannel chan fileInfo, pFileFilterList []string, pSearchContent bool) filepath.WalkFunc {

	fnFileScanner := NewFileScanner(pNamePattern, pContentPattern, pSearchContent)
	fnFileNameFilter := NewListContains(pFileFilterList)

	return func(pFilePath string, pInfo os.FileInfo, pErr error) error {
		if pErr == nil {
			if fnFileNameFilter(pInfo.Name()) {
				if pInfo.IsDir() {
					fInfo := fileInfo{pFilePath, true, false, false, false}
					index := pNamePattern.FindStringIndex(pInfo.Name())
					if index != nil {
						fInfo.foundPathMatch = true
					}
					pOutFileInfoChannel <- fInfo
				} else {
					fInfos := fnFileScanner(pFilePath)
					for _, fInfo := range fInfos {
						pOutFileInfoChannel <- fInfo
					}
				}
			}
		}
		return nil
	}
}

func NewListContains(pFileNameMatchList []string) filterFile {

	if pFileNameMatchList == nil || len(pFileNameMatchList) == 0 {
		return func(pName string) bool { return true }
	}

	return func(pName string) bool {
		for _, namePattern := range pFileNameMatchList {

			isMatch, _ := filepath.Match(namePattern, pName)

			if isMatch {
				return true
			}
		}
		return false
	}
}

func NewFileScanner(pNamePattern *regexp.Regexp, pContentPattern *regexp.Regexp, pContentSearchEnabled bool) fileScanner {

	fnFileNameMatcher := NewStringFinder(pNamePattern)
	fnFileContentMatcher := NewContentFinder(pContentPattern)
	fnZipScanner := NewZipFileScanner(fnFileNameMatcher, fnFileContentMatcher, pContentSearchEnabled)
	fnDefaultScanner := NewNormalFileScanner(fnFileNameMatcher, fnFileContentMatcher, pContentSearchEnabled)

	return func(pFilePath string) []fileInfo {
		processedFiles := fnZipScanner(pFilePath)
		if len(processedFiles) == 0 {
			processedFiles = fnDefaultScanner(pFilePath)
		}
		return processedFiles
	}
}

func NewZipFileScanner(pFnFileNameMatcher stringFinder, pFnFileContentMatcher contentFinder, pContentSearchEnabled bool) fileScanner {
	return func(pFilePath string) []fileInfo {

		fInfoList := make([]fileInfo, 0)

		zipFile, err := zip.OpenReader(pFilePath)

		if err != nil {
			return fInfoList
		}

		fInfoList = append(fInfoList, fileInfo{pFilePath, false, true, false, false})

		if pFnFileNameMatcher(pFilePath) {
			fInfoList[0].foundPathMatch = true
		}

		for _, zFile := range zipFile.File {

			zFileInfo := fileInfo{pFilePath + "@" + zFile.Name, zFile.FileInfo().IsDir(), true, false, false}

			if pFnFileNameMatcher(zFile.Name) {
				zFileInfo.foundPathMatch = true
			}

			if pContentSearchEnabled {
				zFileReader, _ := zFile.Open()

				if pFnFileContentMatcher(zFileReader) {
					zFileInfo.foundContentMatch = true
				}

				zFileReader.Close()
			}

			fInfoList = append(fInfoList, zFileInfo)
		}

		zipFile.Close()

		return fInfoList
	}
}

func NewNormalFileScanner(pFnFileNameMatcher stringFinder, pFnFileContentMatcher contentFinder, pContentSearchEnabled bool) fileScanner {
	return func(pFilePath string) []fileInfo {

		fInfo := make([]fileInfo, 1)

		fInfo[0] = fileInfo{path: pFilePath, dir: false, zip: false, foundContentMatch: false, foundPathMatch: false}

		//Match File Name
		if pFnFileNameMatcher(filepath.Base(pFilePath)) {
			fInfo[0].foundPathMatch = true
		}

		if pContentSearchEnabled {
			f, err := os.Open(pFilePath)
			defer f.Close()

			if err == nil && pFnFileContentMatcher(f) {
				fInfo[0].foundContentMatch = true
			}
		}

		return fInfo
	}
}

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
	return func(pReader io.Reader) bool {

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

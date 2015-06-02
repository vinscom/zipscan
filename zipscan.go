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
	"strings"
	"errors"
	"path"
	"runtime"
	"flag"
	"fmt"
	"archive/zip"
	"bufio"
	"io"
	"os"
	"path/filepath"
	"regexp"
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

var directoryToScan string
var patternToSearch string
var contentSearch bool
var filterFilePatterns csvString

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

func init() {
	flag.StringVar(&directoryToScan, "d", ".", "Directory to scan")
	flag.Var(&filterFilePatterns, "f", "File or Directory Name Filter e.g. *.jar,*.zip")
	flag.StringVar(&patternToSearch, "p", ".*", "RegExp pattern to match file name, path or content")
	flag.BoolVar(&contentSearch, "e", false, "Enable content search")
}

func main() {
	
	runtime.GOMAXPROCS(2)
	
	flag.Parse()
	
	fileInfoChannel := make(chan fileInfo)
	done := make(chan bool)
	
	go fileListPrinter(fileInfoChannel,done)
	
	filepath.Walk(directoryToScan,createWalker(patternToSearch,fileInfoChannel,filterFilePatterns,contentSearch))
	
	close(fileInfoChannel)
	
	<- done
}

func fileListPrinter(allFileInfoChannel chan fileInfo,done chan bool) {
	for {
		select {
			case fInfo , ok := <- allFileInfoChannel :
				if ok {
					if fInfo.foundContentMatch || fInfo.foundPathMatch {
						fmt.Println(fInfo.path)
					}
				} else {
					done <- true
					break
				}
		}
	}
}

func createWalker(pattern string, allFileInfoChannel chan fileInfo,filteFilterList []string,enableContentSearch bool) filepath.WalkFunc {

	compiledPattern := regexp.MustCompile(pattern)
	fileScanner := createfileScanner(compiledPattern,enableContentSearch)
	fileNameScanner := createFileFilter(filteFilterList)

	return func(path string, info os.FileInfo, err error) error {		
		if(err == nil){
			if fileNameScanner(info.Name()){
				if info.IsDir() {
					fInfo := fileInfo{path: path, dir: true, zip: false, foundContentMatch: false, foundPathMatch: false}
					p := compiledPattern.FindStringIndex(info.Name())
					if p != nil {
						fInfo.foundPathMatch = true
					}
					allFileInfoChannel <- fInfo
				} else {
					fInfo := fileScanner(path)
					for _ , i := range fInfo {
						allFileInfoChannel <- i
					}
				}
			}
		}
		return nil
	}
}

func createFileFilter(list []string) filterFile {
	
	if(list == nil || len(list) == 0){
		return func (b string) bool { return true }
	} 
	
	return func (b string) bool {
		for _, a := range list {
			
			m, _ := path.Match(a,b)
	        
			if m {
	            return true
	        }
    	}
    	return false
	}
}

func createStringFinder(r *regexp.Regexp) stringFinder {
	return func(path string) bool {
		match := r.FindStringIndex(path)
		if match == nil {
			return false
		}
		return true
	}
}

func createContentFinder(r *regexp.Regexp) contentFinder {
	return func(f io.Reader) bool {
		bReader := bufio.NewReader(f)
		match := r.FindReaderIndex(bReader)
		if match == nil {
			return false
		}
		return true
	}
}

func createfileScanner(pattern *regexp.Regexp,enableContentSearch bool) fileScanner {

	fnFileNameMatcher := createStringFinder(pattern)
	fnFileContentMatcher := createContentFinder(pattern)
	fnZipReader := createZipFileReader(fnFileNameMatcher, fnFileContentMatcher,enableContentSearch)
	fnNormalReader := createNormalFileReader(fnFileNameMatcher, fnFileContentMatcher,enableContentSearch)

	return func(path string) []fileInfo {
		processedFiles := fnZipReader(path)
		if processedFiles == nil {
			processedFiles = fnNormalReader(path)
		}
		return processedFiles
	}
}

func createZipFileReader(fileNameMatcher stringFinder, fileContentMatcher contentFinder,enableContentSearch bool) fileScanner {
	return func(path string) []fileInfo {

		zipFile, err := zip.OpenReader(path)

		if err != nil {
			return nil
		}

		fInfo := make([]fileInfo, 1)

		fInfo[0] = fileInfo{path: path, dir: false, zip: true, foundContentMatch: false, foundPathMatch: false}

		if fileNameMatcher(path) {
			fInfo[0].foundPathMatch = true
		}

		for _, zFile := range zipFile.File {

			zFileInfo := fileInfo{path: path + "@" + zFile.Name, dir: zFile.FileInfo().IsDir(), zip: true, foundContentMatch: false, foundPathMatch: false}

			if fileNameMatcher(zFile.Name) {
				zFileInfo.foundPathMatch = true
			}
			
			if enableContentSearch {
				zFileReader, _ := zFile.Open()
	
				if fileContentMatcher(zFileReader) {
					zFileInfo.foundContentMatch = true
				}
	
				zFileReader.Close()
			}

			fInfo = append(fInfo, zFileInfo)
		}

		zipFile.Close()

		return fInfo
	}
}

func createNormalFileReader(fileNameMatcher stringFinder, fileContentMatcher contentFinder,enableContentSearch bool) fileScanner {
	return func(fPath string) []fileInfo {

		fInfo := make([]fileInfo, 1)

		fInfo[0] = fileInfo{path: fPath, dir: false, zip: false, foundContentMatch: false, foundPathMatch: false}

		//Match File Name
		if fileNameMatcher(path.Base(fPath)) {
			fInfo[0].foundPathMatch = true
		}
		
		if enableContentSearch {
			nFile, err := os.Open(fPath)
			defer nFile.Close()
	
			if err == nil && fileContentMatcher(nFile) {
				fInfo[0].foundContentMatch = true
			}
		}

		return fInfo
	}
}

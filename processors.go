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
	"os"
	"fmt"
	"path/filepath"
)

//Find all files & dir
func NewFileInfoGenerator(dir string) processorFn {
	return func(f fileInfo, out fileInfoChannel){
		filepath.Walk(dir, func(pFilePath string, pInfo os.FileInfo, pErr error) error {
			if pErr == nil {
				out <- fileInfo{false, pFilePath, pInfo.Name(), pInfo.IsDir(), false, false}
			}
			return nil
		})
	}
}

//Filter files & dir
func NewFileNameFilter(pFileNameMatchList []string) processorFn {
	if pFileNameMatchList == nil || len(pFileNameMatchList) == 0 {
		return func(f fileInfo, out fileInfoChannel){
			out <- f
		}
	}
	
	return func(f fileInfo, out fileInfoChannel) {
		for _, namePattern := range pFileNameMatchList {

			isMatch, _ := filepath.Match(namePattern, f.name)

			if isMatch {
				out <- f
				break
			}
		}
	}
}

//Scan Zip File
func NewZipFileScanner(pFnFileNameMatcher stringFinder, pFnFileContentMatcher contentFinder, pContentSearchEnabled bool) processorFn {
	return func(f fileInfo, out fileInfoChannel) {

		fInfoList := make([]fileInfo, 0)

		zipFile, err := zip.OpenReader(f.path)
		
		if err != nil {
			out <- f
			return
		}

		f.processed = true
		fInfoList = append(fInfoList, f)
		
		if pFnFileNameMatcher(f.name) {
			fInfoList[0].foundPathMatch = true
		}

		for _, zFile := range zipFile.File {

			zFileInfo := fileInfo{false,f.path + "@" + zFile.FileInfo().Name(),zFile.FileInfo().Name(),zFile.FileInfo().IsDir(),false,false}

			if pFnFileNameMatcher(zFile.Name) {
				zFileInfo.foundPathMatch = true
			}

			if pContentSearchEnabled {
				zFileReader, _ := zFile.Open()

				if pFnFileContentMatcher(zFileReader) {
					zFileInfo.foundContentMatch = true
				}

			}

			zFileInfo.processed = true
			fInfoList = append(fInfoList, zFileInfo)
		}

		defer zipFile.Close()
		
		for _,f := range fInfoList {
			out <- f
		}
	}
}

//Scan Noraml File
func NewNormalFileScanner(pFnFileNameMatcher stringFinder, pFnFileContentMatcher contentFinder, pContentSearchEnabled bool) processorFn {
	return func(f fileInfo, out fileInfoChannel) {

		//Match File Name
		if pFnFileNameMatcher(f.name) {
			f.foundPathMatch = true
		}

		if pContentSearchEnabled {
			file, err := os.Open(f.path)
			defer file.Close()

			if err == nil && pFnFileContentMatcher(file) {
				f.foundContentMatch = true
			}
		}
		
		f.processed = true
		
		out <- f
	}
}

//ConsolePrintProcessor
func PrintToConsole(f fileInfo, out fileInfoChannel) {
	if(f.foundContentMatch || f.foundPathMatch){
		fmt.Println(f.name)
	}
}

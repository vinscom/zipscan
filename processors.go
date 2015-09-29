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
func NewTraverseDirectoryProcessor(dir string) processorFn {
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
func NewFileNameFilterProcessor(pFileNameMatchList []string) processorFn {
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

func NewMatchFileNameProcessor(pFnFileNameMatcher stringFinder,pEnable bool) processorFn {
	if pEnable {
		return func(f fileInfo, out fileInfoChannel) {
			if pFnFileNameMatcher(f.name) {
				f.foundPathMatch = true
			}
			out <- f
			
		}
	} else {
		return func(f fileInfo, out fileInfoChannel) {
			out <- f
			
		}
	}
}

//Scan Zip File
func NewZipFileProcessor(pFnFileContentMatcher contentFinder, pContentSearchEnabled bool) processorFn {
	return func(f fileInfo, out fileInfoChannel) {

		fInfoList := make([]fileInfo, 0)

		zipFile, err := zip.OpenReader(f.path)
		
		if err != nil {
			out <- f
			return
		}

		f.processed = true
		fInfoList = append(fInfoList, f)

		for _, zFile := range zipFile.File {

			zFileInfo := fileInfo{false,f.path + "@" + zFile.FileHeader.Name,zFile.FileInfo().Name(),zFile.FileInfo().IsDir(),false,false}

			if pContentSearchEnabled {
				zFileReader, _ := zFile.Open()

				if pFnFileContentMatcher(zFileReader) {
					zFileInfo.foundContentMatch = true
				}

			}

			zFileInfo.processed = true
			fInfoList = append(fInfoList, zFileInfo)
		}

		zipFile.Close()
		
		for _,f := range fInfoList {
			out <- f
		}
	}
}

//Scan Noraml File
func NewNormalFileProcessor(pFnFileContentMatcher contentFinder, pContentSearchEnabled bool) processorFn {
	return func(f fileInfo, out fileInfoChannel) {

		if f.processed || !pContentSearchEnabled {
			out <- f
			return
		}
		
		file, err := os.Open(f.path)
		defer file.Close()

		if err == nil && pFnFileContentMatcher(file) {
			f.foundContentMatch = true
		} 
		
		f.processed = true
		out <- f
	}
}

//ConsolePrintProcessor
func PrintToConsoleProcessor(f fileInfo, out fileInfoChannel) {
	if(f.foundContentMatch || f.foundPathMatch){
		fmt.Println(f.path)
	}
}

package main

import (
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

var directoryToScan string
var patternToSearch string
var enableContentSearch bool

func init() {
	flag.StringVar(&directoryToScan, "d", ".", "Directory to scan")
	flag.StringVar(&patternToSearch, "p", ".*", "RegExp pattern to match file name or path")
	flag.BoolVar(&enableContentSearch, "e", false, "Enable content search")
}

func main() {
	
	runtime.GOMAXPROCS(2)
	
	flag.Parse()
	
	fileInfoChannel := make(chan fileInfo)
	
	go fileListPrinter(fileInfoChannel)
	
	filepath.Walk(directoryToScan,createWalker(patternToSearch,fileInfoChannel))
	
	close(fileInfoChannel)
}

func fileListPrinter(allFileInfoChannel chan fileInfo) {
	for {
		select {
			case fInfo , ok := <- allFileInfoChannel :
				if ok {
					if fInfo.foundContentMatch || fInfo.foundPathMatch {
						fmt.Println(fInfo.path)
					}
				} else {
					break
				}
		}
	}
}

func createWalker(pattern string, allFileInfoChannel chan fileInfo) filepath.WalkFunc {

	compiledPattern := regexp.MustCompile(pattern)
	fileScanner := createfileScanner(compiledPattern)

	return func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			fInfo := fileInfo{path: path, dir: true, zip: false, foundContentMatch: false, foundPathMatch: false}
			p := compiledPattern.FindStringIndex(path)
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
		
		return nil
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

func createfileScanner(pattern *regexp.Regexp) fileScanner {

	fnFileNameMatcher := createStringFinder(pattern)
	fnFileContentMatcher := createContentFinder(pattern)
	fnZipReader := createZipFileReader(fnFileNameMatcher, fnFileContentMatcher)
	fnNormalReader := createNormalFileReader(fnFileNameMatcher, fnFileContentMatcher)

	return func(path string) []fileInfo {
		processedFiles := fnZipReader(path)
		if processedFiles == nil {
			processedFiles = fnNormalReader(path)
		}
		return processedFiles
	}
}

func createZipFileReader(fileNameMatcher stringFinder, fileContentMatcher contentFinder) fileScanner {
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

func createNormalFileReader(fileNameMatcher stringFinder, fileContentMatcher contentFinder) fileScanner {
	return func(path string) []fileInfo {

		fInfo := make([]fileInfo, 1)

		fInfo[0] = fileInfo{path: path, dir: false, zip: false, foundContentMatch: false, foundPathMatch: false}

		if fileNameMatcher(path) {
			fInfo[0].foundPathMatch = true
		}
		
		if enableContentSearch {
			nFile, err := os.Open(path)
			defer nFile.Close()
	
			if err == nil && fileContentMatcher(nFile) {
				fInfo[0].foundContentMatch = true
			}
		}

		return fInfo
	}
}

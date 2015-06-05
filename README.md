# zipscan
Searching for file or content of file in Linux or Mac OS X is difficult task when file is inside compressed file. This utility not only helps you in searching for file and content inside directory, but, it can search for content inside compressed zip files like zip, jar files.

## Usage of ./zipscan:
*  -c="^$": Regular expression we are looking in file 
*  -d=".": Directory to scan (Symbolic Links are not followed)
*  -f=[]: Comma seperated list of patterns of name. Only matching names will be search for content or name
*  -p=".*": Regular expression of name of file or directory
*  -s=false: Enable content search. If this is enabled then Content and File Name patterns become same

## Example
 ./zipscan -f "*.md,*.go" -c "READ" -s true

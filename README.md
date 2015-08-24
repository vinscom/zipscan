# zipscan
Searching for file or content of file in Linux or Mac OS X is difficult task when file is inside compressed file. This utility not only helps you in searching for file and content inside directory, but, it can search for content inside compressed zip files like zip, jar files.

## Usage of zipscan:
*  -c string
    	Regular expression we are looking in file  (default "^$")
*  -d string
    	Directory to scan (Symbolic Links are not followed) (default ".")
*  -f value
    	Comma seperated list of patterns of name. Only matching names will be search for content or name (default [])
*  -p string
    	Regular expression of name of file or directory (default ".*")
*  -s	Enable content search. To eanble add -s=true or -s

## Example
zipscan -f "*.md,*.go" -s -c "READ"

Search for READ in content of file type .md,.go

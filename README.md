# zipscan
Searching for file or content of file in Linux or Mac OS X is difficult task when file is inside compressed file. This utility not only helps you in searching for file and content inside directory, but, it can search for content inside compressed zip files like zip, jar files.

## Usage of zipscan:
*  -d=".": Directory to scan
*  -e=false: Enable content search
*  -p=".*": RegExp pattern to match file path and content

## Example
./zipscan -d . -e true -p Apache

Above example will scan current directory for word "Apache" in name of all files and content of all files including compressed files.

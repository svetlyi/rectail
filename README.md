# RecTail - recursive tail

## Installation:

```
git clone https://github.com/svetlyi/rectail
cd rectail
make
```
## Features

* watches all files recursively in specified directories
* detects new files and removes removed ones

## Help
```
>$ rectail --help
Usage of rectail:
  -delay int
    	delay between files scanning (default 600)
  -log_prefix string
    	just a log file prefix stored in OS temp folder (default "rectail.log")
  -max_offset int
    	max offset from the end of the file for the first printing (default 400)
  -regexps_to_watch value

    	Regular expressions to match files and directories.
    	The order should be the same as start_with entities.
    	If there is no regular expression for the start_with[i]
    	entity, any folders/files (entities) will be added to be
    	watched later.

    	To avoid troubles of replacing special symbols like "." by
    	current folder, enclose your regular expressions in quotes, for example ".*".

    	Example:
    	rectail -start_with /foo/bar -start_with foo -regexps_to_watch "[0-9]+\.log"

  -start_with value

    	Start with directories (which directories to scan).

    	Example:
    	rectail -start_with /foo/bar -start_with foo
```

## Specifics

* does not use inotify as it's not very reliable, especially with recursive entities. It might be 
added in the future to detect special cases, like removing files and then creating with 
the same name (in this case the tool won't detect the change).
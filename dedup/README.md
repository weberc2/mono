# README

`dedup` finds duplicate files in a directory and replaces them with hard-links.
Because proving two files to be duplicates requires fully scanning each file,
and full scans are slow for large files, `dedup` will first rule out files which
*are not* duplicates such as hard links to other files in the directory, files
with unique sizes, files with unique first and last blocks, etc.

It will incrementally deduplicate files and log its progress (in contrast to
`rdfind` which at the time of this writing, did not log its progress).
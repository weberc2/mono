#ifndef SCANNER_H
#define SCANNER_H

#include <stddef.h>

#include "core/io/reader.h"
#include "core/io/writer.h"
#include "core/str/str.h"
#include "core/result/result.h"

typedef struct scan_result
{
    str data;
    error err;
} scan_result;

typedef struct scanner
{
    reader source;
    str buffer;
    str delim;

    // buffer_cursor tells us where to resume reading from after we encounter a
    // delimiter.
    size_t buffer_cursor;

    // last_read_size tells us how much data we last read into the buffer.
    size_t last_read_size;

    // delim_cursor tells us how much of the delimiter we've matched in a given
    // frame. it's only used in cases where a frame ends with some strict
    // prefix of a delimiter such that in the subsequent frame we can tell
    // whether there was a delimiter saddling the two frames or whether there
    // was no delimiter at all. For example, if the delimiter is `abcd` and the
    // first frame ends with `abc`, `delim_cursor` will be 2 to tell us that
    // we matched up to the `c` character in the delimiter. when we read in the
    // next frame of data, we just need to resume matching from
    // `delim[delim_cursor:]` to determine whether or not we've matched a
    // delimiter.
    size_t delim_cursor;

    // end_of_section indicates whether or not we've reached the end of a
    // section.
    bool end_of_section;

    error err;
} scanner;

#define SCANNER(src, buf, del) \
    (scanner)                  \
    {                          \
        .source = (src),       \
        .buffer = (buf),       \
        .delim = (del),        \
    }

typedef struct scanner_new_result
{
    bool ok;
    scanner scanner;
} scanner_new_result;

scanner_new_result scanner_new(reader source, str buffer, str delim);
scan_result scanner_next_frame(scanner *s);
error scanner_next_section(scanner *s);
size_t scanner_write_to(scanner *s, writer dst, result *res);

error ERR_EOF;

#endif // SCANNER_H
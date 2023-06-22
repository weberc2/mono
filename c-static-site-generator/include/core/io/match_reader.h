#ifndef MATCH_READER_H
#define MATCH_READER_H

#include "core/str/str.h"
#include "core/result/result.h"
#include "buffered_reader.h"

typedef struct
{
    buffered_reader *source;
    str match;
    size_t match_cursor;
    bool found_match;
} match_reader;

match_reader match_reader_new(buffered_reader *source, str match);
size_t match_reader_read(match_reader *mr, str buf, result *res);
reader match_reader_to_reader(match_reader *mr);

#endif // MATCH_READER_H
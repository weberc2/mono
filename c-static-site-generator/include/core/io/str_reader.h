#ifndef STR_READER_H
#define STR_READER_H

#include "core/str/str.h"
#include "reader.h"

typedef struct
{
    str buffer;
    size_t cursor;
} str_reader;

#define STR_READER(buf) \
    (str_reader) { .buffer = buf, .cursor = 0 }

str_reader str_reader_new(str buffer);
void str_reader_init(str_reader *sr, str buffer);
size_t str_reader_read(str_reader *sr, str buffer);
io_result str_reader_io_read(str_reader *sr, str buffer);
reader str_reader_to_reader(str_reader *sr);

#endif // STR_READER_H
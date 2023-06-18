#ifndef STR_READER_H
#define STR_READER_H

#include "str/str.h"
#include "reader.h"

typedef struct
{
    str buffer;
    size_t cursor;
} str_reader;

void str_reader_init(str_reader *sr, str buffer);
size_t str_reader_read(str_reader *sr, str buffer);
void str_reader_to_reader(str_reader *sr, reader *out);

#endif // STR_READER_H
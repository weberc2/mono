#ifndef BUFFERED_READER_H
#define BUFFERED_READER_H

#include "str/str.h"
#include "reader.h"
#include "io_result.h"

typedef struct
{
    reader source;
    str buffer;
    size_t cursor;
} buffered_reader;

void buffered_reader_init(buffered_reader *br, reader source, str buf);
size_t buffered_reader_read(buffered_reader *br, str buf, io_result *res);

#endif // BUFFERED_READER_H
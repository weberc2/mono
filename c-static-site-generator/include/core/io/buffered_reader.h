#ifndef BUFFERED_READER_H
#define BUFFERED_READER_H

#include "core/str/str.h"
#include "core/result/result.h"
#include "reader.h"

typedef struct
{
    reader source;
    str buffer;
    size_t cursor;
} buffered_reader;

void buffered_reader_init(buffered_reader *br, reader source, str buf);
size_t buffered_reader_read(buffered_reader *br, str buf, result *res);

#endif // BUFFERED_READER_H
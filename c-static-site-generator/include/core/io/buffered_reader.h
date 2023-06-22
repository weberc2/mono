#ifndef BUFFERED_READER_H
#define BUFFERED_READER_H

#include "core/str/str.h"
#include "core/result/result.h"
#include "reader.h"
#include "writer.h"

typedef struct
{
    reader source;
    str buffer;
    size_t cursor;
    size_t read_end;
} buffered_reader;

void buffered_reader_init(buffered_reader *br, reader source, str buf);
size_t buffered_reader_read(buffered_reader *br, str buf, result *res);
bool buffered_reader_find(
    buffered_reader *br,
    writer w,
    result *res,
    str match);
void buffered_reader_to_reader(buffered_reader *br, reader *r);

#endif // BUFFERED_READER_H
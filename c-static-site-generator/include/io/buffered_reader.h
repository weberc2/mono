#ifndef BUFFERED_READER_H
#define BUFFERED_READER_H

#include "byteslice/byteslice.h"
#include "error/error.h"
#include "reader.h"

typedef struct
{
    reader source;
    byteslice buffer;
    size_t cursor;
} buffered_reader;

void buffered_reader_init(buffered_reader *br, reader source, byteslice buf);
size_t buffered_reader_peek(buffered_reader *br, byteslice buf, errors *errs);
size_t buffered_reader_read(buffered_reader *br, byteslice buf, errors *errs);

#endif // BUFFERED_READER_H
#ifndef WRITER_H
#define WRITER_H

#include "byteslice/byteslice.h"
#include "error/error.h"

typedef size_t (*write_func)(void *, byteslice, errors *);

typedef struct
{
    void *data;
    write_func write;
} writer;

void writer_init(writer *w, void *data, write_func write);
void writer_from_bytestring(writer *w, bytestring *bs);
size_t writer_write(writer w, byteslice bs, errors *errs);

#endif // WRITER_H
#ifndef READER_H
#define READER_H

#include "byteslice/byteslice.h"
#include "error/error.h"

#include <stddef.h>

typedef size_t (*read_func)(void *, byteslice, errors *errs);

typedef struct
{
    void *data;
    read_func read;
} reader;

void reader_init(reader *r, void *data, read_func read);
size_t reader_read(reader r, byteslice bs, errors *errs);

#endif // READER_H
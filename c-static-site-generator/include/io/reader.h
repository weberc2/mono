#ifndef READER_H
#define READER_H

#include "str/str.h"
#include "io_result.h"

#include <stddef.h>

typedef size_t (*read_func)(void *, str, io_result *);

typedef struct
{
    void *data;
    read_func read;
} reader;

void reader_init(reader *r, void *data, read_func read);
size_t reader_read(reader r, str s, io_result *res);

#endif // READER_H
#ifndef READER_H
#define READER_H

#include "core/str/str.h"
#include "core/result/result.h"

#include <stddef.h>

typedef size_t (*read_func)(void *, str, result *);

typedef struct
{
    void *data;
    read_func read;
} reader;

void reader_init(reader *r, void *data, read_func read);
size_t reader_read(reader r, str s, result *res);

#endif // READER_H
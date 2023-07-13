#ifndef READER_H
#define READER_H

#include <stddef.h>

#include "core/str/str.h"
#include "io_result.h"

typedef io_result (*read_func)(void *, str);

typedef struct
{
    void *data;
    read_func read;
} reader;

#define READER(d, r) \
    (reader) { .data = (void *)(d), .read = (read_func)(r) }

void reader_init(reader *r, void *data, read_func read);
reader reader_new(void *data, read_func read);
io_result reader_read(reader r, str s);

#endif // READER_H
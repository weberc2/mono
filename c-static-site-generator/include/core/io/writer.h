#ifndef WRITER_H
#define WRITER_H

#include "core/str/str.h"
#include "core/io/io_result.h"

typedef io_result (*write_func)(void *, str);

typedef struct
{
    void *data;
    write_func write;
} writer;

#define WRITER(d, w) \
    (writer) { .data = (void *)(d), .write = (write_func)(w) }

void writer_init(writer *w, void *data, write_func write);
writer writer_new(void *data, write_func write);
io_result writer_write(writer w, str s);

error ERR_SHORT_WRITE;
error ERR_INVALID_WRITE;

#endif // WRITER_H
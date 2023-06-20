#ifndef WRITER_H
#define WRITER_H

#include "core/str/str.h"
#include "io_result.h"

typedef size_t (*write_func)(void *, str, io_result *);

typedef struct
{
    void *data;
    write_func write;
} writer;

void writer_init(writer *w, void *data, write_func write);
size_t writer_write(writer w, str s, io_result *res);

#endif // WRITER_H
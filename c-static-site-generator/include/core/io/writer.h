#ifndef WRITER_H
#define WRITER_H

#include "core/str/str.h"
#include "core/result/result.h"

typedef size_t (*write_func)(void *, str, result *);

typedef struct
{
    void *data;
    write_func write;
} writer;

void writer_init(writer *w, void *data, write_func write);
size_t writer_write(writer w, str s, result *res);

#endif // WRITER_H
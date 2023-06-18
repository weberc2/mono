#ifndef WRITER_H
#define WRITER_H

#include "str/str.h"
#include "error/error.h"

typedef size_t (*write_func)(void *, str, errors *);

typedef struct
{
    void *data;
    write_func write;
} writer;

void writer_init(writer *w, void *data, write_func write);
void writer_from_string(writer *w, string *s);
size_t writer_write(writer w, str s, errors *errs);

#endif // WRITER_H
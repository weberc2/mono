#include "io/reader.h"

void reader_init(reader *r, void *data, read_func read)
{
    r->data = data;
    r->read = read;
}

size_t reader_read(reader r, byteslice bs, errors *errs)
{
    return r.read(r.data, bs, errs);
}
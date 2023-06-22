#include "core/io/reader.h"
#include "core/str/str.h"

void reader_init(reader *r, void *data, read_func read)
{
    r->data = data;
    r->read = read;
}

reader reader_new(void *data, read_func read)
{
    reader r;
    reader_init(&r, data, read);
    return r;
}

size_t reader_read(reader r, str bs, result *res)
{
    return r.read(r.data, bs, res);
}
#include "core/io/writer.h"
#include "core/result/result.h"

void __attribute__((constructor)) init()
{
    error_const(&ERR_SHORT_WRITE, "short write");
    error_const(&ERR_INVALID_WRITE, "invalid write");
}

void writer_init(writer *w, void *data, write_func write)
{
    w->data = data;
    w->write = write;
}

size_t writer_write(writer w, str s, result *res)
{
    return w.write(w.data, s, res);
}
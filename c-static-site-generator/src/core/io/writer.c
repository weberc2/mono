#include "core/io/writer.h"
#include "core/io/io_result.h"

void writer_init(writer *w, void *data, write_func write)
{
    w->data = data;
    w->write = write;
}

size_t writer_write(writer w, str s, io_result *res)
{
    return w.write(w.data, s, res);
}
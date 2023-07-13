#include "core/io/writer.h"

static void __attribute__((constructor)) init()
{
    error_const(&ERR_SHORT_WRITE, "short write");
    error_const(&ERR_INVALID_WRITE, "invalid write");
}

void writer_init(writer *w, void *data, write_func write)
{
    w->data = data;
    w->write = write;
}

writer writer_new(void *data, write_func write)
{
    writer w;
    writer_init(&w, data, write);
    return w;
}

io_result writer_write(writer w, str s)
{
    return w.write(w.data, s);
}
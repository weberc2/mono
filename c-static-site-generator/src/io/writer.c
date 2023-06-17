#include "io/writer.h"

void writer_init(writer *w, void *data, write_func write)
{
    w->data = data;
    w->write = write;
}

size_t writer_write(writer w, byteslice bs, errors *errs)
{
    return w.write(w.data, bs, errs);
}

size_t bytestring_write(bytestring *bs, byteslice buf, errors *errs)
{
    bytestring_push_slice(bs, buf);
    return buf.len;
}

void writer_from_bytestring(writer *w, bytestring *bs)
{
    w->data = bs;
    w->write = (write_func)bytestring_write;
}
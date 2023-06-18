#include "io/writer.h"

void writer_init(writer *w, void *data, write_func write)
{
    w->data = data;
    w->write = write;
}

size_t writer_write(writer w, str s, errors *errs)
{
    return w.write(w.data, s, errs);
}

size_t string_write(string *s, str buf, errors *errs)
{
    string_push_slice(s, buf);
    return buf.len;
}

void writer_from_string(writer *w, string *s)
{
    w->data = s;
    w->write = (write_func)string_write;
}
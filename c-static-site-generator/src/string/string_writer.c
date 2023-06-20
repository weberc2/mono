#include "string/string_writer.h"
#include "string/string.h"
#include "io/writer.h"
#include "io/io_result.h"

size_t string_write(string *s, str buf, io_result *res)
{
    io_result_ok(res);
    string_push_slice(s, buf);
    return buf.len;
}

void string_writer(writer *w, string *s)
{
    writer_init(w, s, (write_func)string_write);
}
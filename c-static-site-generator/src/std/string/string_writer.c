#include "std/string/string_writer.h"
#include "std/string/string.h"
#include "core/io/writer.h"
#include "core/result/result.h"

size_t string_write(string *s, str buf, result *res)
{
    result_ok(res);
    string_push_slice(s, buf);
    return buf.len;
}

void string_writer(writer *w, string *s)
{
    writer_init(w, s, (write_func)string_write);
}
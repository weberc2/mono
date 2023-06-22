#include "std/string/string_writer.h"
#include "std/string/string.h"
#include "core/io/writer.h"
#include "core/result/result.h"

size_t string_write(string *s, str buf, result *res)
{
    *res = result_ok();
    string_push_slice(s, buf);
    return buf.len;
}

writer string_writer(string *s)
{
    return writer_new(s, (write_func)string_write);
}
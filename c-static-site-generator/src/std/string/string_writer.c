#include "std/string/string_writer.h"
#include "std/string/string.h"
#include "core/io/writer.h"

io_result string_write(string *s, str buf)
{
    string_push_slice(s, buf);
    return IO_RESULT_OK(buf.len);
}

writer string_writer(string *s)
{
    return writer_new(s, (write_func)string_write);
}
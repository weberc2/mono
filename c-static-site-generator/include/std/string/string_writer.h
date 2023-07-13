#ifndef STRING_WRITER_H
#define STRING_WRITER_H

#include "std/string/string.h"
#include "core/io/writer.h"

#define STRING_WRITER(s)                   \
    (writer)                               \
    {                                      \
        .data = s,                         \
        .write = (write_func)string_write, \
    }

io_result string_write(string *s, str buf);
writer string_writer(string *s);

#endif // STRING_WRITER_H
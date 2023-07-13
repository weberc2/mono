#include "core/io/copy.h"
#include "core/fmt/fmt_fprintf.h"
#include "core/fmt/fmt_reader.h"

io_result fmt_fprintf_buf(writer w, str format, fmt_args args, str buf)
{
    fmt_reader fr = fmt_reader_new(format, args);
    reader r = fmt_reader_to_reader(&fr);
    return copy_buf(w, r, buf);
}

io_result fmt_fprintf(writer w, str format, fmt_args args)
{
    return fmt_fprintf_buf(w, format, args, STR_ARR((char[256]){0}));
}
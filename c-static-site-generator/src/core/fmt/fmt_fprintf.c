#include "core/io/copy.h"
#include "core/fmt/fmt_fprintf.h"
#include "core/fmt/fmt_reader.h"

fmt_result fmt_fprintf_buf(writer w, str format, fmt_args args, str buf)
{
    fmt_reader fr = fmt_reader_new(format, args);
    reader r = fmt_reader_to_reader(&fr);
    result res;
    size_t nc = copy_buf(w, r, buf, &res);
    return (fmt_result){.ok = res.ok, .size = nc, .err = res.err};
}

fmt_result fmt_fprintf(writer w, str format, fmt_args args)
{
    return fmt_fprintf_buf(w, format, args, STR_ARR((char[256]){0}));
}
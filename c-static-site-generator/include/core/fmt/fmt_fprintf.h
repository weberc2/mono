#ifndef FMT_FPRINTF_H
#define FMT_FPRINTF_H

#include <stddef.h>
#include "core/error/error.h"
#include "core/io/writer.h"
#include "core/str/str.h"
#include "fmt_arg.h"

typedef struct
{
    bool ok;
    size_t size;
    error err;
} fmt_result;

#define FMT_FPRINTF(w, fmt, ...) \
    fmt_fprintf((w), STR((fmt)), FMT_ARGS(__VA_ARGS__))

fmt_result fmt_fprintf_buf(writer w, str format, fmt_args args, str buf);
fmt_result fmt_fprintf(writer w, str format, fmt_args args);
#endif // FMT_FPRINTF_H
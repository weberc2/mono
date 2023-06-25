#ifndef FMT_READER_H
#define FMT_READER_H

#include <stddef.h>
#include <stdbool.h>
#include "core/str/str.h"
#include "core/result/result.h"
#include "core/panic/panic.h"
#include "core/io/reader.h"
#include "core/io/writer.h"
#include "fmt_arg.h"

typedef struct
{
    str format;
    fmt_args args;
    size_t cursor;
    bool reading_arg;
} fmt_reader;

fmt_reader fmt_reader_new(str format, fmt_args args);
size_t fmt_reader_read(fmt_reader *fr, str buf, result *res);
reader fmt_reader_to_reader(fmt_reader *fr);

#endif // FMT_READER_H
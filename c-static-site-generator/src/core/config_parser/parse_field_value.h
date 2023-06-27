#ifndef PARSE_FIELD_VALUE_H
#define PARSE_FIELD_VALUE_H

#include <stddef.h>
#include "core/error/error.h"
#include "core/io/reader.h"
#include "core/io/writer.h"
#include "parse_status.h"

typedef struct parse_field_value_result
{
    parse_status tag;
    union
    {
        size_t buffer_position;
        error io_err;
        struct
        {
        } match_failure;
    } result;
} parse_field_value_result;

#define PARSE_FIELD_VALUE_OK(bp)        \
    (parse_field_value_result)          \
    {                                   \
        .tag = parse_ok,                \
        .result.buffer_position = (bp), \
    }

#define PARSE_FIELD_VALUE_IO_ERROR(e) \
    (parse_field_value_result)        \
    {                                 \
        .tag = parse_io_error,        \
        .result.io_err = (e),         \
    }

#define PARSE_FIELD_VALUE_MATCH_FAILURE \
    (parse_field_value_result)          \
    {                                   \
        .tag = parse_match_failure,     \
        .result.match_failure = {},     \
    }

parse_field_value_result parse_field_value(
    reader r,
    writer w,
    str buf,
    size_t cursor,
    size_t last_read_end);

#endif // PARSE_FIELD_VALUE_H
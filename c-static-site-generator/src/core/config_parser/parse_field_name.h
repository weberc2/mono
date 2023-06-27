#ifndef PARSE_FIELD_NAME_H
#define PARSE_FIELD_NAME_H

#include "core/str/str.h"
#include "core/io/reader.h"
#include "parse_status.h"
#include "field.h"

typedef struct parse_field_name_result
{
    parse_status tag;
    union
    {
        struct
        {
            size_t buffer_position;
            field_handle field_handle;
        } ok;
        error io_err;
        struct
        {
        } match_failure;
    } result;
} parse_field_name_result;

#define PARSE_FIELD_NAME_OK(fh, bp)      \
    (parse_field_name_result)            \
    {                                    \
        .tag = parse_ok,                 \
        .result.ok.field_handle = fh,    \
        .result.ok.buffer_position = bp, \
    }

#define PARSE_FIELD_NAME_MATCH_FAILURE \
    (parse_field_name_result)          \
    {                                  \
        .tag = parse_match_failure,    \
        .result.match_failure = {},    \
    }

#define PARSE_FIELD_NAME_IO_ERROR(e) \
    (parse_field_name_result)        \
    {                                \
        .tag = parse_io_error,       \
        .result.io_err = (e),        \
    }

parse_field_name_result parse_field_name(
    reader r,
    fields fields,
    str buf,
    size_t cursor,
    size_t last_read_end);

#endif // PARSE_FIELD_NAME_H
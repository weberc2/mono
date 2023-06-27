#ifndef FIELD_PARSER_H
#define FIELD_PARSER_H

#include <stdbool.h>
#include "core/panic/panic.h"
#include "core/str/str.h"
#include "core/io/reader.h"
#include "core/io/writer.h"

typedef struct field
{
    str name;
    writer dst;
    bool match_failed;
} field;

field field_new(str name, writer dst);

#define FIELD(n, d, mf)     \
    (field)                 \
    {                       \
        .name = (n),        \
        .dst = d,           \
        .match_failed = mf, \
    }

typedef struct fields
{
    field *data;
    size_t len;
} fields;

#define FIELDS(...)                                            \
    (fields)                                                   \
    {                                                          \
        .data = (field[]){__VA_ARGS__},                        \
        .len = sizeof((field[]){__VA_ARGS__}) / sizeof(field), \
    }

bool fields_has_valid(fields fields);

typedef enum parse_status
{
    parse_ok,
    parse_io_error,
    parse_match_failure,
} parse_status;

static inline str parse_status_str(parse_status status)
{
    switch (status)
    {
    case parse_ok:
        return STR_LIT("PARSE_STATUS_OK");
    case parse_io_error:
        return STR_LIT("PARSE_STATUS_IO_ERROR");
    case parse_match_failure:
        return STR_LIT("PARSE_STATUS_MATCH_FAILURE");
    default:
        panic("invalid parse status: `%d`", status);
    }
}

typedef size_t field_handle;

typedef struct fields_match_result
{
    bool match;
    size_t buffer_position;
    size_t field_handle;
} fields_match_result;

fields_match_result fields_match_name(
    fields fields,
    size_t field_name_cursor,
    str buf);

#define FIELDS_MATCH_OK(fh, bp)  \
    (fields_match_result)        \
    {                            \
        .match = true,           \
        .buffer_position = (bp), \
        .field_handle = (fh),    \
    }

#define FIELDS_MATCH_FAILURE  \
    (fields_match_result)     \
    {                         \
        .match = false,       \
        .buffer_position = 0, \
        .field_handle = 0,    \
    }

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

typedef struct parse_field_value_result
{
    parse_status tag;
    size_t total_size;
    size_t buffer_position;
    error err;
} parse_field_value_result;

#define PARSE_FIELD_VALUE_RESULT_OK(ts, bp) \
    (parse_field_value_result)              \
    {                                       \
        .tag = parse_ok,                    \
        .total_size = (ts),                 \
        .buffer_position = (bp),            \
        .err = ERROR_NULL,                  \
    }

parse_field_value_result parse_field_value(
    reader r,
    writer w,
    str buf,
    size_t cursor,
    size_t last_read_end);

typedef struct parse_result
{
    parse_status tag;
    error io_err;
} parse_result;

parse_result parse_result_ok();
parse_result parse_result_io_error(error io_err);
parse_result parse_result_match_failure();

parse_result parse_field(reader r, fields fields, str buf);

#endif // FIELD_PARSER_H
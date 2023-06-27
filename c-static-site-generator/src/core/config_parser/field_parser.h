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

typedef struct field_match_result
{
    parse_status tag;
    size_t buffer_position;
    field_handle field_handle;
    error io_err;
} field_match_result;

#define FIELD_MATCH_RESULT_SUCCESS(fh, bp) \
    (field_match_result)                   \
    {                                      \
        .tag = parse_ok,                   \
        .field_handle = fh,                \
        .buffer_position = bp,             \
    }

#define FIELD_MATCH_RESULT_FAILURE  \
    (field_match_result)            \
    {                               \
        .tag = parse_match_failure, \
        .field_handle = 0,          \
        .buffer_position = 0,       \
        .io_err = ERROR_NULL,       \
    }

field_match_result fields_match_name(
    fields fields,
    size_t field_name_cursor,
    str buf);

field_match_result parse_field_name(reader r, fields fields, str buf);

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
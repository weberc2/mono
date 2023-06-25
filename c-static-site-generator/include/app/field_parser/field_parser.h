#ifndef FIELD_PARSER_H
#define FIELD_PARSER_H

#include <stdbool.h>
#include "core/str/str.h"
#include "core/io/reader.h"
#include "std/string/string.h"

typedef struct field
{
    str name;
    string data;
    bool match_failed;
} field;

field field_new(str name);

#define STRING(lit)             \
    (string)                    \
    {                           \
        .data = lit,            \
        .cap = sizeof(lit) - 1, \
        .len = sizeof(lit) - 1, \
    }

#define FIELD(n, d, mf)     \
    (field)                 \
    {                       \
        .name = (n),        \
        .data = d,          \
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

typedef size_t field_handle;

typedef struct field_match_result
{
    bool match;
    size_t buffer_position;
    field_handle field_handle;
    result io_error;
} field_match_result;

#define FIELD_MATCH_RESULT_SUCCESS(fh, bp) \
    (field_match_result)                   \
    {                                      \
        .match = true,                     \
        .field_handle = fh,                \
        .buffer_position = bp,             \
    }

#define FIELD_MATCH_RESULT_FAILURE \
    (field_match_result)           \
    {                              \
        .match = false,            \
        .field_handle = 0,         \
        .buffer_position = 0,      \
        .io_error = RESULT_OK,     \
    }

field_match_result fields_match_name(
    fields fields,
    size_t field_name_cursor,
    str buf);

field_match_result parse_field_name(reader r, fields fields, str buf);

typedef struct parse_field_value_result
{
    bool ok;
    size_t total_size;
    size_t buffer_position;
    error err;
} parse_field_value_result;

#define PARSE_FIELD_VALUE_RESULT_OK(ts, bp) \
    (parse_field_value_result)              \
    {                                       \
        .ok = true,                         \
        .total_size = (ts),                 \
        .buffer_position = (bp),            \
        .err = ERROR_NULL,                  \
    }

parse_field_value_result parse_field_value(reader r, writer w, str buf);
#endif // FIELD_PARSER_H
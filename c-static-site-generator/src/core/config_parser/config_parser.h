#ifndef FIELD_PARSER_H
#define FIELD_PARSER_H

#include <stdbool.h>
#include "core/panic/panic.h"
#include "core/str/str.h"
#include "core/io/reader.h"
#include "core/io/writer.h"
#include "parse_status.h"
#include "field.h"

typedef struct parse_field_result
{
    parse_status tag;
    union
    {
        struct
        {
        } ok;
        struct
        {
        } match_failure;
        error io_err;
        field_handle field;
    } result;
} parse_field_result;

#define PARSE_FIELD_OK   \
    (parse_field_result) \
    {                    \
        .tag = parse_ok, \
        .result.ok = {}, \
    }

#define PARSE_FIELD_IO_ERROR(e) \
    (parse_field_result)        \
    {                           \
        .tag = parse_io_error,  \
        .result.io_err = (e),   \
    }

#define PARSE_FIELD_MATCH_FAILURE   \
    (parse_field_result)            \
    {                               \
        .tag = parse_match_failure, \
        .result.match_failure = {}, \
    }

typedef struct parse_result
{
    parse_status tag;
    error io_err;
} parse_result;

typedef struct config_parser
{
    reader reader;
    fields fields;
    str buffer;
    size_t cursor;
    size_t last_read_end;
} config_parser;

parse_field_result config_parser_parse_field(config_parser *parser);

#endif // FIELD_PARSER_H
#ifndef PARSE_TO_FIELDS_H
#define PARSE_TO_FIELDS_H

#include "config_parser.h"
#include "field.h"

typedef enum config_parser_parse_to_fields_status
{
    config_parser_parse_to_fields_status_ok,
    config_parser_parse_to_fields_status_io_error,
    config_parser_parse_to_fields_status_parse_error,
} config_parser_parse_to_fields_status;

static inline str config_parser_parse_to_fields_status_to_str(
    config_parser_parse_to_fields_status s)
{
    switch (s)
    {
    case config_parser_parse_to_fields_status_ok:
        return STR_LIT("CONFIG_PARSER_PARSE_TO_FIELDS_STATUS_OK");
    case config_parser_parse_to_fields_status_io_error:
        return STR_LIT("CONFIG_PARSER_PARSE_TO_FIELDS_STATUS_IO_ERROR");
    case config_parser_parse_to_fields_status_parse_error:
        return STR_LIT("CONFIG_PARSER_PARSE_TO_FIELDS_STATUS_PARSE_ERROR");
    }
}

typedef struct config_parser_parse_to_fields_result
{
    config_parser_parse_to_fields_status status;
    error io_err;
} config_parser_parse_to_fields_result;

config_parser_parse_to_fields_result config_parser_parse_to_fields(
    config_parser *parser,
    fields *fields);

#define CONFIG_PARSER_PARSE_TO_FIELDS_OK                   \
    (config_parser_parse_to_fields_result)                 \
    {                                                      \
        .status = config_parser_parse_to_fields_status_ok, \
        .io_err = ERROR_NULL,                              \
    }

#define CONFIG_PARSER_PARSE_TO_FIELDS_IO_ERROR(e)                \
    (config_parser_parse_to_fields_result)                       \
    {                                                            \
        .status = config_parser_parse_to_fields_status_io_error, \
        .io_err = (e),                                           \
    }

#define CONFIG_PARSER_PARSE_TO_FIELDS_PARSE_ERROR                   \
    (config_parser_parse_to_fields_result)                          \
    {                                                               \
        .status = config_parser_parse_to_fields_status_parse_error, \
        .io_err = ERROR_NULL,                                       \
    }

#endif // PARSE_TO_FIELDS_H
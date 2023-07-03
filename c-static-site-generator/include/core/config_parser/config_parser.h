#ifndef CONFIG_PARSER_H
#define CONFIG_PARSER_H

#include "core/str/str.h"
#include "core/io/reader.h"
#include "core/io/writer.h"

// transitions
// current state | event      | next state
// --------------+------------+------------
// beginning     | value_next | beginning
// beginning     | key_next   | parsing key
// parsing key   | ':'        | parsed key
// parsing key   | key_next   | parsing key
// parsing key   | value_next | parsing key
// parsing key   | '\n'       | parse error
// parsed key    | key_next   | parsed key
// parsed key    | value_next | parsing value
// parsing value | value_next | parsing value
// parsing value | key_next   | parsing value
// parsing value | '\n'       | parsed value
// parsed value  | value_next | parsed value
// parsed value  | key_next   | parsing key
typedef enum config_parser_state
{
    config_parser_state_start,
    config_parser_state_parsing_key,
    config_parser_state_parsed_key,
    config_parser_state_parsing_value,
    config_parser_state_parsed_value,
    config_parser_state_parse_error,
    config_parser_state_io_error,
    config_parser_state_eof,
} config_parser_state;

static inline bool config_parser_state_is_error(config_parser_state state)
{
    return state == config_parser_state_eof ||
           state == config_parser_state_io_error ||
           state == config_parser_state_parse_error;
}

static inline str config_parser_state_to_str(config_parser_state state)
{
    switch (state)
    {
    case config_parser_state_start:
        return STR_LIT("CONFIG_PARSER_STATE_START");
    case config_parser_state_parsing_key:
        return STR_LIT("CONFIG_PARSER_STATE_PARSING_KEY");
    case config_parser_state_parsed_key:
        return STR_LIT("CONFIG_PARSER_STATE_PARSED_KEY");
    case config_parser_state_parsing_value:
        return STR_LIT("CONFIG_PARSER_STATE_PARSING_VALUE");
    case config_parser_state_parsed_value:
        return STR_LIT("CONFIG_PARSER_STATE_PARSED_VALUE");
    case config_parser_state_parse_error:
        return STR_LIT("CONFIG_PARSER_STATE_PARSE_ERROR");
    case config_parser_state_io_error:
        return STR_LIT("CONFIG_PARSER_STATE_IO_ERROR");
    case config_parser_state_eof:
        return STR_LIT("CONFIG_PARSER_STATE_EOF");
    }
}

typedef struct config_parser_result
{
    config_parser_state state;
    str bytes;
    error io_err;
} config_parser_result;

#define CONFIG_PARSER_START                 \
    (config_parser_result)                  \
    {                                       \
        .state = config_parser_state_start, \
        .bytes = STR_EMPTY,                 \
        .io_err = ERROR_NULL,               \
    }

#define CONFIG_PARSER_PARSED_KEY(bs)             \
    (config_parser_result)                       \
    {                                            \
        .state = config_parser_state_parsed_key, \
        .bytes = (bs),                           \
        .io_err = ERROR_NULL,                    \
    }

#define CONFIG_PARSER_PARSING_KEY(bs)             \
    (config_parser_result)                        \
    {                                             \
        .state = config_parser_state_parsing_key, \
        .bytes = (bs),                            \
        .io_err = ERROR_NULL,                     \
    }

#define CONFIG_PARSER_PARSED_VALUE(bs)             \
    (config_parser_result)                         \
    {                                              \
        .state = config_parser_state_parsed_value, \
        .bytes = (bs),                             \
        .io_err = ERROR_NULL,                      \
    }

#define CONFIG_PARSER_PARSING_VALUE(bs)             \
    (config_parser_result)                          \
    {                                               \
        .state = config_parser_state_parsing_value, \
        .bytes = (bs),                              \
        .io_err = ERROR_NULL,                       \
    }

#define CONFIG_PARSER_EOF(bs)             \
    (config_parser_result)                \
    {                                     \
        .state = config_parser_state_eof, \
        .bytes = (bs),                    \
        .io_err = ERROR_NULL,             \
    }

#define CONFIG_PARSER_IO_ERROR(bs, e)          \
    (config_parser_result)                     \
    {                                          \
        .state = config_parser_state_io_error, \
        .bytes = (bs),                         \
        .io_err = (e),                         \
    }

#define CONFIG_PARSER_PARSE_ERROR                 \
    (config_parser_result)                        \
    {                                             \
        .state = config_parser_state_parse_error, \
        .bytes = STR_EMPTY,                       \
        .io_err = ERROR_NULL,                     \
    }

typedef struct config_parser
{
    reader source;
    str buffer;
    size_t cursor;
    size_t last_read_size;
    config_parser_state state;
    error io_err;
} config_parser;

config_parser config_parser_new(reader source, str buffer);
config_parser_result config_parser_key_next(config_parser *parser);
config_parser_result config_parser_value_next(config_parser *parser);
size_t config_parser_key_write_to(
    config_parser *parser,
    writer w,
    result *res);
size_t config_parser_value_write_to(
    config_parser *parser,
    writer w,
    result *res);

#endif // CONFIG_PARSER_H
#include <stdint.h>
#include "core/config_parser/config_parser.h"

config_parser config_parser_new(reader source, str buffer)
{
    return (config_parser){
        .source = source,
        .buffer = buffer,
        .cursor = 0,
        .last_read_size = 0,
        .state = config_parser_state_start,
        .io_err = ERROR_NULL,
    };
}

static inline str config_parser_data(config_parser *parser)
{
    return str_slice(parser->buffer, parser->cursor, parser->last_read_size);
}

static inline bool config_parser_refresh(config_parser *parser)
{
    result res = result_new();
    parser->last_read_size = reader_read(parser->source, parser->buffer, &res);
    parser->cursor = 0;
    if (!res.ok)
    {
        parser->state = config_parser_state_io_error;
        parser->io_err = res.err;
        return false;
    }
    if (parser->last_read_size < 1)
    {
        parser->state = config_parser_state_eof;
        return false;
    }
    return true;
}

config_parser_result config_parser_to_result(config_parser *parser)
{
    return (config_parser_result){
        .state = parser->state,
        .bytes = config_parser_data(parser),
        .io_err = parser->io_err,
    };
}

bool config_parser_skip_whitespace(
    config_parser *parser,
    bool including_newlines)
{
    while (true)
    {
        str data = config_parser_data(parser);
        for (size_t i = 0; i < data.len; i++)
        {
            if (data.data[i] != ' ' &&
                data.data[i] != '\t' &&
                (!including_newlines || data.data[i] != '\n'))
            {
                parser->cursor += i;
                return true;
            }
        }

        if (!config_parser_refresh(parser))
        {
            return false;
        }
    }
}

config_parser_result config_parser_key_next(config_parser *parser)
{
    switch (parser->state)
    {
    case config_parser_state_eof:
    case config_parser_state_io_error:
    case config_parser_state_parse_error:
        return (config_parser_result){
            .state = parser->state,
            .bytes = config_parser_data(parser),
        };
    case config_parser_state_parsed_key:
        return CONFIG_PARSER_PARSED_KEY(STR_EMPTY);
    case config_parser_state_parsing_value:
        return CONFIG_PARSER_PARSING_VALUE(STR_EMPTY);
    case config_parser_state_parsed_value:
    case config_parser_state_start:
        if (!config_parser_skip_whitespace(parser, true))
        {
            return config_parser_to_result(parser);
        }
        parser->state = config_parser_state_parsing_key;
        break;
    case config_parser_state_parsing_key:
        break;
    }

    if (parser->cursor >= parser->last_read_size)
    {
        if (!config_parser_refresh(parser))
        {
            return config_parser_to_result(parser);
        }
    }

    str data = config_parser_data(parser);
    for (size_t i = 0; i < data.len; i++)
    {
        if (data.data[i] == '\n')
        {
            parser->state = config_parser_state_parse_error;
            return CONFIG_PARSER_PARSE_ERROR;
        }

        if (data.data[i] == ':')
        {
            config_parser_result res = CONFIG_PARSER_PARSED_KEY(
                str_slice(parser->buffer, parser->cursor, parser->cursor + i));
            parser->cursor += i + 1; // at worst `cursor == last_read_size`
            parser->state = config_parser_state_parsed_key;
            return res;
        }
    }

    // if we got here, then the whole `data` slice is free of delimiters; we
    // can return it to the caller as-is and without changing the parser state.
    parser->cursor += data.len;
    return CONFIG_PARSER_PARSING_KEY(data);
}

config_parser_result config_parser_value_next(config_parser *parser)
{
    switch (parser->state)
    {
    case config_parser_state_eof:
    case config_parser_state_io_error:
    case config_parser_state_parse_error:
        return (config_parser_result){
            .state = parser->state,
            .bytes = config_parser_data(parser),
        };
    case config_parser_state_parsed_value:
        return CONFIG_PARSER_PARSED_VALUE(STR_EMPTY);
    case config_parser_state_start:
        return CONFIG_PARSER_START;
    case config_parser_state_parsing_key:
        return CONFIG_PARSER_PARSING_KEY(STR_EMPTY);
    case config_parser_state_parsed_key:
        if (!config_parser_skip_whitespace(parser, false))
        {
            return config_parser_to_result(parser);
        }
        parser->state = config_parser_state_parsing_value;
        break;
    case config_parser_state_parsing_value:
        break;
    }

    if (parser->cursor >= parser->last_read_size)
    {
        if (!config_parser_refresh(parser))
        {
            return config_parser_to_result(parser);
        }
    }

    str data = config_parser_data(parser);
    for (size_t i = 0; i < data.len; i++)
    {
        if (data.data[i] == '\n')
        {
            config_parser_result res = CONFIG_PARSER_PARSED_VALUE(
                str_slice(data, 0, i));
            parser->cursor += i + 1; // at worst `cursor == last_read_size`
            parser->state = config_parser_state_parsed_value;
            return res;
        }

        // don't allow malicious input to break shitty C programs
        if (data.data[i] == '\0')
        {
            const uint8_t ASCII_SUBSTITUTE_CHARACTER = 0x1A;
            data.data[i] = ASCII_SUBSTITUTE_CHARACTER;
        }
    }

    // if we got here, then the whole `data` slice is free of delimiters; we
    // can return it to the caller as-is and without changing the parser state.
    parser->cursor += data.len;
    return CONFIG_PARSER_PARSING_VALUE(data);
}

size_t config_parser_write_to_helper(
    config_parser *parser,
    writer w,
    result *res,
    config_parser_state continue_state,
    config_parser_result (*next)(config_parser *))
{
    size_t total_written = 0;
    while (true)
    {
        config_parser_result parse_result = next(parser);
        size_t nr = writer_write(w, parse_result.bytes, res);
        total_written += nr;
        if (nr != parse_result.bytes.len)
        {
            *res = result_err(ERR_SHORT_WRITE);
            return total_written;
        }
        if (parse_result.state != continue_state)
        {
            return total_written;
        }
    }
}

size_t config_parser_key_write_to(
    config_parser *parser,
    writer w,
    result *res)
{
    return config_parser_write_to_helper(
        parser,
        w,
        res,
        config_parser_state_parsing_key,
        config_parser_key_next);
}

size_t config_parser_value_write_to(
    config_parser *parser,
    writer w,
    result *res)
{
    return config_parser_write_to_helper(
        parser,
        w,
        res,
        config_parser_state_parsing_value,
        config_parser_value_next);
}
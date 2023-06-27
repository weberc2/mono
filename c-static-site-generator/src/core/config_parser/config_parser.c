#include "config_parser.h"
#include "parse_field_name.h"
#include "parse_field_value.h"

parse_field_result config_parser_parse_field(config_parser *parser)
{
    parse_field_name_result name_res = parse_field_name(
        parser->reader,
        parser->fields,
        parser->buffer,
        parser->cursor,
        parser->last_read_end);

    switch (name_res.tag)
    {
    case parse_io_error:
        return PARSE_FIELD_IO_ERROR(name_res.result.io_err);
    case parse_match_failure:
        return PARSE_FIELD_MATCH_FAILURE;
    case parse_ok:
        break;
    }

    parser->cursor = name_res.result.ok.buffer_position;
    field f = parser->fields.data[name_res.result.ok.field_handle];

    parse_field_value_result value_res = parse_field_value(
        parser->reader,
        f.dst,
        parser->buffer,
        parser->cursor,
        parser->last_read_end);

    switch (value_res.tag)
    {
    case parse_io_error:
        return PARSE_FIELD_IO_ERROR(value_res.result.io_err);
    case parse_match_failure:
        return PARSE_FIELD_MATCH_FAILURE;
    case parse_ok:
        break;
    }

    parser->cursor = value_res.result.buffer_position;
    return PARSE_FIELD_OK;
}

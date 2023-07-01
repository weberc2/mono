#include "core/config_parser/parse_to_fields.h"
#include "core/panic/panic.h"

void fields_reset(fields *fields)
{
    fields->cursor = 0;
    for (size_t i = 0; i < fields->len; i++)
    {
        if (fields->data[i].status == field_status_disqualified)
        {
            fields->data[i].status = field_status_inconclusive;
        }
    }
}

field *fields_try_match(fields *fields, str buf, bool end)
{
    for (size_t i = 0; i < fields->len; i++)
    {
        field *f = &fields->data[i];
        if (f->status != field_status_inconclusive)
        {
            continue;
        }

        str key_remaining = str_slice(
            f->key,
            fields->cursor,
            f->key.len);

        if (str_eq(key_remaining, buf))
        {
            if (end)
            {
                f->status = field_status_matched;
                return f;
            }
            continue;
        }

        if (!str_has_prefix(key_remaining, buf))
        {
            f->status = field_status_disqualified;
            continue;
        }
    }

    fields->cursor += buf.len;
    return NULL;
}

typedef enum field_match
{
    field_match_success,
    field_match_failed,
    field_match_eof,
    field_match_io_error,
    field_match_parse_error,
} field_match;

typedef struct field_match_result
{
    field_match type;
    field *match;
    error io_err;
} field_match_result;

field_match_result fields_match_key(fields *fields, config_parser *parser)
{
    while (true)
    {
        config_parser_result res = config_parser_key_next(parser);
        field *f;
        switch (res.state)
        {
        case config_parser_state_eof:
            return (field_match_result){
                .type = field_match_eof,
                .match = NULL,
                .io_err = ERROR_NULL,
            };
        case config_parser_state_io_error:
            return (field_match_result){
                .type = field_match_io_error,
                .match = NULL,
                .io_err = res.io_err,
            };
        case config_parser_state_parse_error:
            return (field_match_result){
                .type = field_match_parse_error,
                .match = NULL,
                .io_err = ERROR_NULL,
            };
        case config_parser_state_parsing_key:
            if (fields_try_match(fields, res.bytes, false) != NULL)
            {
                panic(
                    "program error: fields_try_match() returned non-NULL "
                    "field before the config_parser finished parsing the key");
            }
            break;
        case config_parser_state_parsed_key:
            f = fields_try_match(fields, res.bytes, true);
            if (f == NULL)
            {
                return (field_match_result){
                    .type = field_match_failed,
                    .match = NULL,
                    .io_err = ERROR_NULL,
                };
            }
            return (field_match_result){
                .type = field_match_success,
                .match = f,
                .io_err = ERROR_NULL,
            };
        default:
            panic(
                "program error: illegal state change: "
                "(CONFIG_PARSER_STATE_PARSING_KEY|CONFIG_PARSER_STATE_START)"
                " -> %s",
                config_parser_state_to_str(res.state));
        }
    }
}

config_parser_parse_to_fields_result config_parser_parse_to_fields(
    config_parser *parser,
    fields *fields)
{
    while (true)
    {
        fields_reset(fields);
        field_match_result res = fields_match_key(fields, parser);
        result io_res = result_new();
        switch (res.type)
        {
        case field_match_success:
            config_parser_value_write_to(
                parser,
                res.match->value,
                &io_res);
            if (!io_res.ok)
            {
                return CONFIG_PARSER_PARSE_TO_FIELDS_IO_ERROR(io_res.err);
            }
            continue;
        case field_match_failed:
            continue;
        case field_match_eof:
            return CONFIG_PARSER_PARSE_TO_FIELDS_OK;
        case field_match_io_error:
            return CONFIG_PARSER_PARSE_TO_FIELDS_IO_ERROR(res.io_err);
        case field_match_parse_error:
            return CONFIG_PARSER_PARSE_TO_FIELDS_PARSE_ERROR;
        }
    }
}
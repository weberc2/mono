#include "core/io/str_reader.h"

#include "test_helpers.h"
#include "field_parser.h"

#define LIT_READER(lit)                            \
    (reader)                                       \
    {                                              \
        .data = (void *)&STR_READER(STR_LIT(lit)), \
        .read = (read_func)str_reader_io_read,     \
    }

typedef struct parse_field_value_test
{
    char *name;
    reader input;
    str buf;
    size_t cursor;
    size_t last_read_end;
    str wanted_data;
    parse_field_value_result wanted_result;
} parse_field_value_test;

parse_field_value_test parse_field_value_tests[] = {
    {
        .name = "test_parse_field_value:empty",
        .input = LIT_READER(""),
        .buf = STR_ARR((char[8]){0}),
        .cursor = 0,
        .last_read_end = 0,
        .wanted_data = STR_LIT(""),
        .wanted_result = PARSE_FIELD_VALUE_RESULT_OK(0, 0),
    },
    {
        .name = "test_parse_field_value:eof",
        .input = LIT_READER("hello"),
        .buf = STR_ARR((char[8]){0}),
        .cursor = 0,
        .last_read_end = 0,
        .wanted_data = STR_LIT("hello"),
        .wanted_result = PARSE_FIELD_VALUE_RESULT_OK(5, 5),
    },
    {
        .name = "test_parse_field_value:input-ends-with-newline",
        .input = LIT_READER("hello\n"),
        .buf = STR_ARR((char[8]){0}),
        .cursor = 0,
        .last_read_end = 0,
        .wanted_data = STR_LIT("hello"),
        .wanted_result = PARSE_FIELD_VALUE_RESULT_OK(5, 5),
    },
    {
        .name = "test_parse_field_value:newline-in-middle-of-input",
        .input = LIT_READER("hello\nworld"),
        .buf = STR_ARR((char[8]){0}),
        .cursor = 0,
        .last_read_end = 0,
        .wanted_data = STR_LIT("hello"),
        .wanted_result = PARSE_FIELD_VALUE_RESULT_OK(5, 5),
    },
    {
        .name = "test_parse_field_value:multi-iterations-to-find-newline",
        .input = LIT_READER("hello world\ngreetings"),
        .buf = STR_ARR((char[3]){0}),
        .cursor = 0,
        .last_read_end = 0,
        .wanted_data = STR_LIT("hello world"),
        .wanted_result = PARSE_FIELD_VALUE_RESULT_OK(11, 1),
    },
    {
        .name = "test_parse_field_value:search-initial-buffer-first",
        .input = LIT_READER(" world\ngreetings"),
        .buf = STR_ARR((char[21]){"OLDDATA:hello:BADDATA"}),
        .cursor = 8,
        .last_read_end = 13,
        .wanted_data = STR_LIT("hello world"),
        .wanted_result = PARSE_FIELD_VALUE_RESULT_OK(11, 6),
    },
};

bool parse_field_value_test_run(parse_field_value_test *tc)
{
    test_init(tc->name);
    string found_data = string_new();
    parse_field_value_result found_result = parse_field_value(
        tc->input,
        string_writer(&found_data),
        tc->buf,
        tc->cursor,
        tc->last_read_end);
    if (tc->wanted_result.tag != found_result.tag)
    {
        return test_fail(
            "result.ok: wanted `%s`; found `%s`",
            parse_status_str(tc->wanted_result.tag).data,
            parse_status_str(found_result.tag).data);
    }

    if (tc->wanted_result.tag != parse_ok)
    {
        if (tc->wanted_result.buffer_position != found_result.buffer_position)
        {
            return test_fail(
                "result.buffer_position: wanted `%zu`; found `%zu`",
                tc->wanted_result.buffer_position,
                found_result.buffer_position);
        }

        if (tc->wanted_result.total_size != found_result.total_size)
        {
            return test_fail(
                "result.total_size: wanted `%zu`; found `%zu`",
                tc->wanted_result.total_size,
                found_result.total_size);
        }
    }

    if (!str_eq(tc->wanted_data, string_borrow(&found_data)))
    {
        char w[256] = {0}, f[256] = {0};
        string_copy_to_c(f, &found_data, sizeof(f));
        str_copy_to_c(w, tc->wanted_data, sizeof(w));
        return test_fail(
            "data: wanted `%s` (len: %zu); found `%s` (len: %zu)",
            w,
            tc->wanted_data.len,
            f,
            found_data.len);
    }

    return test_success();
}

bool test_parse_field_value()
{
    for (
        size_t i = 0;
        i < sizeof(parse_field_value_tests) / sizeof(parse_field_value_test);
        i++)
    {
        if (!parse_field_value_test_run(&parse_field_value_tests[i]))
        {
            return false;
        }
    }

    return true;
}
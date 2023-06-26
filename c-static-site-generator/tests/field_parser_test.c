
#include "core/io/str_reader.h"
#include "std/string/string.h"
#include "std/string/string_writer.h"
#include "core/config_parser/field_parser.h"
#include "test.h"

typedef struct match_name_test
{
    char *name;
    fields fields;
    size_t field_name_cursor;
    str buf;
    fields wanted_fields;
    field_match_result wanted_result;
} match_name_test;

#define EMPTY_STRING_WRITER STRING_WRITER(&STRING_NEW)

match_name_test match_name_tests[] = {
    {
        .name = "test_fields_match_name:no-matches",
        .fields = FIELDS(FIELD(STR_LIT("hello"), EMPTY_STRING_WRITER, false)),
        .field_name_cursor = 0,
        .buf = STR_LIT("foobar"),
        .wanted_fields = FIELDS(
            FIELD(STR_LIT("hello"), EMPTY_STRING_WRITER, true)),
        .wanted_result = FIELD_MATCH_RESULT_FAILURE,
    },
    {
        .name = "test_fields_match_name:match-at-buffer-start",
        .fields = FIELDS(FIELD(STR_LIT("hello"), EMPTY_STRING_WRITER, false)),
        .field_name_cursor = 0,
        .buf = STR_LIT("hello:"),
        .wanted_fields = FIELDS(
            FIELD(STR_LIT("hello"), EMPTY_STRING_WRITER, false)),
        .wanted_result = FIELD_MATCH_RESULT_SUCCESS(0, 5),
    },
    {
        .name = "test_fields_match_name:buffer-matches-field-name-minus-delim",
        .fields = FIELDS(FIELD(STR_LIT("hello"), EMPTY_STRING_WRITER, false)),
        .field_name_cursor = 0,
        .buf = STR_LIT("hello"),
        .wanted_fields = FIELDS(
            FIELD(STR_LIT("hello"), EMPTY_STRING_WRITER, false)),
        .wanted_result = FIELD_MATCH_RESULT_FAILURE,
    },
    {
        // this test captures the case where we've already matched `hello` in
        // a previous iteration, but we need to find the delimiter at the start
        // of the next buffer.
        .name = "test_fields_match_name:resuming-match-at-buffer-start",
        .fields = FIELDS(FIELD(STR_LIT("hello"), EMPTY_STRING_WRITER, false)),
        .field_name_cursor = 5,
        .buf = STR_LIT(":foo"),
        .wanted_fields = FIELDS(
            FIELD(STR_LIT("hello"), EMPTY_STRING_WRITER, false)),
        .wanted_result = FIELD_MATCH_RESULT_SUCCESS(0, 0),
    },
    {
        // this test captures the case where we've already matched `hello` in
        // a previous iteration, but we need to find the delimiter at the start
        // of the next buffer; however, the buffer does not start with the
        // delimiter and thus it is not a match.
        .name = "test_fields_match_name:resuming-match-at-buffer-start",
        .fields = FIELDS(FIELD(STR_LIT("hello"), EMPTY_STRING_WRITER, false)),
        .field_name_cursor = 5,
        .buf = STR_LIT("foo"),
        .wanted_fields = FIELDS(
            FIELD(STR_LIT("hello"), EMPTY_STRING_WRITER, true)),
        .wanted_result = FIELD_MATCH_RESULT_FAILURE,
    },
    {
        // test that even though the current buffer matches the post-cursor
        // field name, we still fail the match because the field had already
        // been disqualified.
        .name = "test_fields_match_name:skips-previously-failed-match",
        .fields = FIELDS((field){
            .name = STR_LIT("foohello"),
            .dst = EMPTY_STRING_WRITER,
            .match_failed = true,
        }),
        .field_name_cursor = 3,
        .buf = STR_LIT("hello:"),
        .wanted_fields = FIELDS((field){
            .name = STR_LIT("foohello"),
            .dst = EMPTY_STRING_WRITER,
            .match_failed = true,
        }),
        .wanted_result = FIELD_MATCH_RESULT_FAILURE,
    },
    {
        // if the buffer doesn't completely match a field name, then we return
        // a match failure but we leave the field's `match_failed` field to
        // set to false to indicate that the field may still match successfully
        // in the coming iterations.
        .name = "test_fields_match_name:buffer-is-prefix-of-field-name",
        .fields = FIELDS((field){
            .name = STR_LIT("hello"),
            .dst = EMPTY_STRING_WRITER,
            .match_failed = false,
        }),
        .field_name_cursor = 0,
        .buf = STR_LIT("hell"),
        .wanted_fields = FIELDS((field){
            .name = STR_LIT("hello"),
            .dst = EMPTY_STRING_WRITER,
            .match_failed = false,
        }),
        .wanted_result = FIELD_MATCH_RESULT_FAILURE,
    },
    {
        // if we've previously matched some of the name in a previous
        // iteration, and the current buffer doesn't completely match a field
        // name, then we return a match failure but we leave the field's
        // `match_failed` field to set to false to indicate that the field may
        // still match successfully in the coming iterations.
        .name = "test_fields_match_name:partial-match-middle-of-field-name",
        .fields = FIELDS((field){
            .name = STR_LIT("hello"),
            .dst = EMPTY_STRING_WRITER,
            .match_failed = false,
        }),
        .field_name_cursor = 2,
        .buf = STR_LIT("ll"),
        .wanted_fields = FIELDS((field){
            .name = STR_LIT("hello"),
            .dst = EMPTY_STRING_WRITER,
            .match_failed = false,
        }),
        .wanted_result = FIELD_MATCH_RESULT_FAILURE,
    },
};

bool assert_fields_consistent(fields wanted, fields input)
{
    if (wanted.len != input.len)
    {
        // if we got here, then our test case is invalid--the function under
        // test can't manipulate the number of fields, so it's invalid to
        // expect a different number of fields than are provided.
        return test_fail(
            "test initialization error: mismatching number of input and "
            "desired fields: wanted `%zu` fields but provided `%zu` fields "
            "as input",
            wanted.len,
            input.len);
    }

    return true;
}

#define ASSERT_FIELDS_CONSISTENT(wanted, input)   \
    if (!assert_fields_consistent(wanted, input)) \
    {                                             \
        return false;                             \
    }

bool assert_fields_eq(fields wanted_fields, fields found_fields)
{
    for (size_t i = 0; i < found_fields.len; i++)
    {
        field wanted = wanted_fields.data[i], found = found_fields.data[i];
        if (!str_eq(wanted.name, found.name))
        {
            char w[256] = {0}, f[256] = {0};
            return test_fail(
                "fields[%zu]: name: wanted `%s`; found `%s`",
                i,
                w,
                f);
        }

        if (!str_eq(
                string_borrow((string *)wanted.dst.data),
                string_borrow((string *)found.dst.data)))
        {
            char w[256] = {0}, f[256] = {0};
            return test_fail(
                "fields[%zu]: data: wanted `%s`; found `%s`",
                i,
                w,
                f);
        }

        if (wanted.match_failed != found.match_failed)
        {
            return test_fail(
                "fields[%zu]: match_failed: wanted `%s`; found `%s`",
                i,
                wanted.match_failed ? "true" : "false",
                found.match_failed ? "true" : "false");
        }
    }

    return true;
}

#define ASSERT_FIELDS_EQ(wanted, found)   \
    if (!assert_fields_eq(wanted, found)) \
    {                                     \
        return false;                     \
    }

bool assert_result_eq(result wanted, result found)
{
    if (wanted.ok != found.ok)
    {
        return test_fail(
            "ok: wanted `%s`; found `%s`",
            wanted.ok ? "true" : "false",
            found.ok ? "true" : "false");
    }
    return true;
}

#define ASSERT_RESULT_EQ(wanted, found)   \
    if (!assert_result_eq(wanted, found)) \
    {                                     \
        return false;                     \
    }

bool assert_field_match_result_eq(
    field_match_result wanted,
    field_match_result found)
{
    if (!wanted.match && found.match)
    {
        return test_fail(
            "unexpected match found: field index `%zu`; buffer position `%zu`",
            found.field_handle,
            found.buffer_position);
    }
    if (wanted.match && !found.match)
    {
        return test_fail("expected match but found none");
    }
    if (wanted.field_handle != found.field_handle)
    {
        return test_fail(
            "field_handle: wanted `%zu`; found `%zu`",
            wanted.field_handle,
            found.field_handle);
    }
    if (wanted.buffer_position != found.buffer_position)
    {
        return test_fail(
            "buffer_position: wanted `%zu`; found `%zu`",
            wanted.buffer_position,
            found.buffer_position);
    }
    ASSERT_RESULT_EQ(wanted.io_error, found.io_error);
    return true;
}

#define ASSERT_FIELD_MATCH_RESULT_EQ(wanted, found)   \
    if (!assert_field_match_result_eq(wanted, found)) \
    {                                                 \
        return false;                                 \
    }

void string_writer_fields_drop(fields *fields)
{
    for (size_t i = 0; i < fields->len; i++)
    {
        string_drop((string *)fields->data[i].dst.data);
    }
}

bool match_name_test_run(match_name_test *tc)
{
    test_init(tc->name);
    ASSERT_FIELDS_CONSISTENT(tc->wanted_fields, tc->fields);

    // free any memory allocated in the internal strings
    TEST_DEFER(string_writer_fields_drop, &tc->fields);
    TEST_DEFER(string_writer_fields_drop, &tc->wanted_fields);

    for (size_t i = 0; i < tc->fields.len; i++)
    {
        TEST_DEFER(string_drop, tc->fields.data[i].dst.data);
        TEST_DEFER(string_drop, tc->wanted_fields.data[i].dst.data);
    }

    field_match_result found_result = fields_match_name(
        tc->fields,
        tc->field_name_cursor,
        tc->buf);
    ASSERT_FIELD_MATCH_RESULT_EQ(tc->wanted_result, found_result);
    ASSERT_FIELDS_EQ(tc->wanted_fields, tc->fields);
    return test_success();
}

bool test_fields_match_name()
{
    for (
        size_t i = 0;
        i < sizeof(match_name_tests) / sizeof(match_name_test);
        i++)
    {
        if (!match_name_test_run(&match_name_tests[i]))
        {
            return false;
        }
    }

    return true;
}

typedef struct parse_field_name_test
{
    char *name;
    str input;
    fields fields;
    str buf;
    field_match_result wanted_match_result;
    fields wanted_fields;
} parse_field_name_test;

parse_field_name_test parse_field_name_tests[] = {
    {
        // when no fields match the buffer, expect that all fields are marked
        // `.match_failed = true` and the returned result is marked
        // `.match = false`.
        .name = "test_parse_field_name:no-match",
        .input = STR_LIT("world"),
        .fields = FIELDS((field){
            .name = STR_LIT("hello"),
            .dst = EMPTY_STRING_WRITER,
            .match_failed = false,
        }),
        .buf = STR_ARR((char[32]){0}),
        .wanted_match_result = FIELD_MATCH_RESULT_FAILURE,
        .wanted_fields = FIELDS((field){
            .name = STR_LIT("hello"),
            .dst = EMPTY_STRING_WRITER,
            .match_failed = true,
        }),
    },
    {
        // when all fields are marked `.match_failed = true`, then expect the
        // result is marked `.match = false`.
        .name = "test_parse_field_name:aborts-when-no-fields-match",
        .input = STR_LIT("bar"),
        .fields = FIELDS((field){
            .name = STR_LIT("bar"),
            .dst = EMPTY_STRING_WRITER,
            .match_failed = true,
        }),
        .buf = STR_ARR((char[32]){0}),
        .wanted_match_result = FIELD_MATCH_RESULT_FAILURE,
        .wanted_fields = FIELDS(
            FIELD(STR_LIT("bar"), EMPTY_STRING_WRITER, true)),
    },
    {
        .name = "test_parse_field_name:match-found",
        .input = STR_LIT("foo:bar"),
        .fields = FIELDS(FIELD(STR_LIT("foo"), EMPTY_STRING_WRITER, false)),
        .buf = STR_ARR((char[32]){0}),
        .wanted_match_result = FIELD_MATCH_RESULT_SUCCESS(0, 3),
        .wanted_fields = FIELDS(
            FIELD(STR_LIT("foo"), EMPTY_STRING_WRITER, false)),
    },
    {
        // make sure matching works even when we have to loop multiple times to
        // match a field.
        .name = "test_parse_field_name:multi-iterations-per-match",
        .input = STR_LIT("foo:bar"),
        .fields = FIELDS(FIELD(STR_LIT("foo"), EMPTY_STRING_WRITER, false)),
        .buf = STR_ARR((char[3]){0}),
        .wanted_match_result = FIELD_MATCH_RESULT_SUCCESS(0, 0),
        .wanted_fields = FIELDS(
            FIELD(STR_LIT("foo"), EMPTY_STRING_WRITER, false)),
    },
};

bool parse_field_name_test_run(parse_field_name_test *tc)
{
    test_init(tc->name);
    ASSERT_FIELDS_CONSISTENT(tc->wanted_fields, tc->fields);

    // free any memory allocated in the internal strings
    TEST_DEFER(string_writer_fields_drop, &tc->fields);
    TEST_DEFER(string_writer_fields_drop, &tc->wanted_fields);

    field_match_result found_match_result = parse_field_name(
        str_reader_to_reader(&STR_READER(tc->input)),
        tc->fields,
        tc->buf);
    ASSERT_FIELD_MATCH_RESULT_EQ(tc->wanted_match_result, found_match_result);
    ASSERT_FIELDS_EQ(tc->wanted_fields, tc->fields);
    return test_success();
}

bool test_parse_field_name()
{
    for (
        size_t i = 0;
        i < sizeof(parse_field_name_tests) / sizeof(parse_field_name_test);
        i++)
    {
        if (!parse_field_name_test_run(&parse_field_name_tests[i]))
        {
            return false;
        }
    }

    return true;
}

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
    str wanted_data;
    parse_field_value_result wanted_result;
} parse_field_value_test;

parse_field_value_test parse_field_value_tests[] = {
    {
        .name = "test_parse_field_value:empty",
        .input = LIT_READER(""),
        .buf = STR_ARR((char[8]){0}),
        .wanted_data = STR_LIT(""),
        .wanted_result = PARSE_FIELD_VALUE_RESULT_OK(0, 0),
    },
    {
        .name = "test_parse_field_value:eof",
        .input = LIT_READER("hello"),
        .buf = STR_ARR((char[8]){0}),
        .wanted_data = STR_LIT("hello"),
        .wanted_result = PARSE_FIELD_VALUE_RESULT_OK(5, 5),
    },
    {
        .name = "test_parse_field_value:input-ends-with-newline",
        .input = LIT_READER("hello\n"),
        .buf = STR_ARR((char[8]){0}),
        .wanted_data = STR_LIT("hello"),
        .wanted_result = PARSE_FIELD_VALUE_RESULT_OK(5, 5),
    },
    {
        .name = "test_parse_field_value:newline-in-middle-of-input",
        .input = LIT_READER("hello\nworld"),
        .buf = STR_ARR((char[8]){0}),
        .wanted_data = STR_LIT("hello"),
        .wanted_result = PARSE_FIELD_VALUE_RESULT_OK(5, 5),
    },
    {
        .name = "test_parse_field_value:multi-iterations-to-find-newline",
        .input = LIT_READER("hello world\ngreetings"),
        .buf = STR_ARR((char[3]){0}),
        .wanted_data = STR_LIT("hello world"),
        .wanted_result = PARSE_FIELD_VALUE_RESULT_OK(11, 1),
    },
};

bool parse_field_value_test_run(parse_field_value_test *tc)
{
    test_init(tc->name);
    string found_data = string_new();
    parse_field_value_result found_result = parse_field_value(
        tc->input,
        string_writer(&found_data),
        tc->buf);
    if (tc->wanted_result.ok != found_result.ok)
    {
        return test_fail(
            "result.ok: wanted `%s`; found `%s`",
            tc->wanted_result.ok ? "true" : "false",
            found_result.ok ? "true" : "false");
    }

    if (!tc->wanted_result.ok)
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

bool field_parser_tests()
{
    return test_fields_match_name() &&
           test_parse_field_name() &&
           test_parse_field_value();
}
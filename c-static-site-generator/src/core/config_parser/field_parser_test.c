
#include "core/io/str_reader.h"
#include "std/string/string.h"
#include "std/string/string_writer.h"
#include "core/config_parser/field_parser.h"
#include "core/testing/test.h"

#include "test_helpers.h"

typedef struct match_name_test
{
    char *name;
    fields fields;
    size_t field_name_cursor;
    str buf;
    fields wanted_fields;
    field_match_result wanted_result;
} match_name_test;

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
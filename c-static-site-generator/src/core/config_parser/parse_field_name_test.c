#include "core/io/str_reader.h"
#include "core/testing/test.h"

#include "test_helpers.h"
#include "field_parser.h"

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
        .input = STR("world"),
        .fields = FIELDS((field){
            .name = STR("hello"),
            .dst = EMPTY_STRING_WRITER,
            .match_failed = false,
        }),
        .buf = STR_ARR((char[32]){0}),
        .wanted_match_result = FIELD_MATCH_RESULT_FAILURE,
        .wanted_fields = FIELDS((field){
            .name = STR("hello"),
            .dst = EMPTY_STRING_WRITER,
            .match_failed = true,
        }),
    },
    {
        // when all fields are marked `.match_failed = true`, then expect the
        // result is marked `.match = false`.
        .name = "test_parse_field_name:aborts-when-no-fields-match",
        .input = STR("bar"),
        .fields = FIELDS((field){
            .name = STR("bar"),
            .dst = EMPTY_STRING_WRITER,
            .match_failed = true,
        }),
        .buf = STR_ARR((char[32]){0}),
        .wanted_match_result = FIELD_MATCH_RESULT_FAILURE,
        .wanted_fields = FIELDS(
            FIELD(STR("bar"), EMPTY_STRING_WRITER, true)),
    },
    {
        .name = "test_parse_field_name:match-found",
        .input = STR("foo:bar"),
        .fields = FIELDS(FIELD(STR("foo"), EMPTY_STRING_WRITER, false)),
        .buf = STR_ARR((char[32]){0}),
        .wanted_match_result = FIELD_MATCH_RESULT_SUCCESS(0, 3),
        .wanted_fields = FIELDS(
            FIELD(STR("foo"), EMPTY_STRING_WRITER, false)),
    },
    {
        // make sure matching works even when we have to loop multiple times to
        // match a field.
        .name = "test_parse_field_name:multi-iterations-per-match",
        .input = STR("foo:bar"),
        .fields = FIELDS(FIELD(STR("foo"), EMPTY_STRING_WRITER, false)),
        .buf = STR_ARR((char[3]){0}),
        .wanted_match_result = FIELD_MATCH_RESULT_SUCCESS(0, 0),
        .wanted_fields = FIELDS(
            FIELD(STR("foo"), EMPTY_STRING_WRITER, false)),
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
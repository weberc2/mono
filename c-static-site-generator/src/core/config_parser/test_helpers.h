#ifndef TEST_HELPERS_H
#define TEST_HELPERS_H

#include "core/config_parser/field_parser.h"
#include "core/testing/test.h"
#include "std/string/string_writer.h"

#define EMPTY_STRING_WRITER STRING_WRITER(&STRING_NEW)

static inline bool assert_fields_consistent(fields wanted, fields input)
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

static inline bool assert_fields_eq(fields wanted_fields, fields found_fields)
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

static inline bool assert_result_eq(result wanted, result found)
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

static inline bool assert_field_match_result_eq(
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

static inline void string_writer_fields_drop(fields *fields)
{
    for (size_t i = 0; i < fields->len; i++)
    {
        string_drop((string *)fields->data[i].dst.data);
    }
}

#endif // TEST_HELPERS_H
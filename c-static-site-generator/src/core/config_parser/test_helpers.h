#ifndef TEST_HELPERS_H
#define TEST_HELPERS_H

#include "core/testing/test.h"
#include "std/string/string_writer.h"

#include "fields_match_name.h"
#include "parse_field_name.h"

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

        if (wanted.match_status != found.match_status)
        {
            return test_fail(
                "fields[%zu]: match_failed: wanted `%s`; found `%s`",
                i,
                field_match_status_str(wanted.match_status).data,
                field_match_status_str(found.match_status).data);
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

static inline bool assert_fields_match_result_eq(
    fields_match_result wanted,
    fields_match_result found)
{
    if (wanted.match != found.match)
    {
        return test_fail(
            "expected match `%s`; found `%s`",
            wanted.match ? "true" : "false",
            found.match ? "true" : "false");
    }

    if (wanted.match)
    {
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
    }

    return true;
}

#define ASSERT_FIELDS_MATCH_RESULT_EQ(wanted, found)   \
    if (!assert_fields_match_result_eq(wanted, found)) \
    {                                                  \
        return false;                                  \
    }

static inline bool assert_parse_field_name_result_eq(
    parse_field_name_result wanted,
    parse_field_name_result found)
{
    if (wanted.tag != found.tag)
    {
        return test_fail(
            "expected status `%s`; found `%s`",
            parse_status_str(wanted.tag).data,
            parse_status_str(found.tag).data);
    }

    if (wanted.tag == parse_ok)
    {
        if (wanted.result.ok.field_handle != found.result.ok.field_handle)
        {
            return test_fail(
                "field_handle: wanted `%zu`; found `%zu`",
                wanted.result.ok.field_handle,
                found.result.ok.field_handle);
        }

        if (wanted.result.ok.buffer_position !=
            found.result.ok.buffer_position)
        {
            return test_fail(
                "buffer_position: wanted `%zu`; found `%zu`",
                wanted.result.ok.buffer_position,
                found.result.ok.buffer_position);
        }
    }

    return true;
}

#define ASSERT_PARSE_FIELD_NAME_RESULT_EQ(wanted, found)   \
    if (!assert_parse_field_name_result_eq(wanted, found)) \
    {                                                      \
        return false;                                      \
    }

static inline void string_writer_fields_drop(fields *fields)
{
    for (size_t i = 0; i < fields->len; i++)
    {
        string_drop((string *)fields->data[i].dst.data);
    }
}

#endif // TEST_HELPERS_H
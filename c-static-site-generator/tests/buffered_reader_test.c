#include "core/testing/test.h"
#include "core/io/buffered_reader.h"
#include "core/io/str_reader.h"
#include "core/io/copy.h"
#include "core/panic/panic.h"
#include "std/string/string.h"
#include "std/string/string_formatter.h"
#include "std/string/string_writer.h"

typedef struct
{
    char *name;
    size_t inner_buf_size;
    size_t outer_buf_size;
    char *src;
    size_t src_size;
    char *match;
    size_t match_size;
    char *wanted_prelude;
    size_t wanted_prelude_size;
    char *wanted_postlude;
    size_t wanted_postlude_size;
    bool wanted_match;
    bool wanted_error;
} find_test_case;

static find_test_case test_cases[] = {
    {
        .name = "test_buffered_reader_find:simple",
        .inner_buf_size = 5,
        .outer_buf_size = 2,
        .src = "hello world!",
        .src_size = sizeof("hello world!") - 1,
        .match = "world",
        .match_size = sizeof("world") - 1,
        .wanted_prelude = "hello ",
        .wanted_prelude_size = sizeof("hello ") - 1,
        .wanted_postlude = "!",
        .wanted_postlude_size = sizeof("!") - 1,
        .wanted_match = true,
        .wanted_error = false,
    },
    {
        .name = "test_buffered_reader_find:big_out_buf",
        .inner_buf_size = 5,
        .outer_buf_size = 128,
        .src = "hello world!",
        .src_size = sizeof("hello world!") - 1,
        .match = "world",
        .match_size = sizeof("world") - 1,
        .wanted_prelude = "hello ",
        .wanted_prelude_size = sizeof("hello ") - 1,
        .wanted_postlude = "!",
        .wanted_postlude_size = sizeof("!") - 1,
        .wanted_match = true,
        .wanted_error = false,
    },
    {
        .name = "test_buffered_reader_find:big_inner_buf",
        .inner_buf_size = 128,
        .outer_buf_size = 5,
        .src = "hello world!",
        .src_size = sizeof("hello world!") - 1,
        .match = "world",
        .match_size = sizeof("world") - 1,
        .wanted_prelude = "hello ",
        .wanted_prelude_size = sizeof("hello ") - 1,
        .wanted_postlude = "!",
        .wanted_postlude_size = sizeof("!") - 1,
        .wanted_match = true,
        .wanted_error = false,
    },
};

bool find_test_case_run(find_test_case *tc)
{
    test_init(tc->name);

    // init strs and bufs
    str src = str_new(tc->src, tc->src_size);
    str match = str_new(tc->match, tc->match_size);
    str inner_buf = STR_ARR((char[256]){0});
    str wanted_prelude = str_new(tc->wanted_prelude, tc->wanted_prelude_size);
    str wanted_postlude = str_new(
        tc->wanted_postlude,
        tc->wanted_postlude_size);

    // init reader
    reader r = str_reader_to_reader(&STR_READER(src));
    buffered_reader br = buffered_reader_new(r, inner_buf);

    // init writer
    string s = string_new();
    writer w = string_writer(&s);
    result res = result_new();

    bool found = buffered_reader_find(&br, w, match, &res);

    if (tc->wanted_match && !found)
    {
        return test_fail(
            "expected to find `%s` in `%s`, but failed",
            tc->match,
            tc->src);
    }

    if (!tc->wanted_match && found)
    {
        return test_fail(
            "unexpectedly found `%s` in `%s`",
            tc->match,
            tc->src);
    }

    if (res.ok && tc->wanted_error)
    {
        return test_fail("expected error but found ok");
    }

    if (!res.ok && !tc->wanted_error)
    {
        char m[256] = {0};
        return test_fail(
            "unexpected error: %s",
            error_to_raw(res.err, m, sizeof(m)));
    }

    str actual_prelude = string_borrow(&s);
    if (!str_eq(wanted_prelude, actual_prelude))
    {
        char wanted[256] = {0}, actual[256] = {0};
        str_copy_to_c(wanted, wanted_prelude, sizeof(wanted));
        str_copy_to_c(actual, actual_prelude, sizeof(actual));

        return test_fail("prelude: wanted `%s`; found `%s`", wanted, actual);
    }

    string postlude = string_new();
    w = string_writer(&postlude);
    buffered_reader_to_reader(&br, &r);
    copy(w, r, &res);
    str actual_postlude = string_borrow(&postlude);
    ASSERT_OK(IO_RESULT_ERR(res.err));
    if (!str_eq(wanted_postlude, actual_postlude))
    {
        char wanted[256] = {0}, actual[256] = {0};
        str_copy_to_c(wanted, wanted_postlude, sizeof(wanted));
        str_copy_to_c(actual, actual_postlude, sizeof(actual));

        return test_fail("postlude: wanted `%s`; found `%s`", wanted, actual);
    }

    return test_success();
}

bool buffered_reader_find_tests()
{
    for (size_t i = 0; i < sizeof(test_cases) / sizeof(find_test_case); i++)
    {
        if (!find_test_case_run(&test_cases[i]))
        {
            return false;
        }
    }

    return true;
}
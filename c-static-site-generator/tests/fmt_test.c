#include "core/fmt/fmt_reader.h"
#include "core/fmt/fmt_fprintf.h"
#include "std/string/string.h"
#include "std/string/string_writer.h"
#include "std/string/string_formatter.h"
#include "core/testing/test.h"

#define ARR(...) __VA_ARGS__, sizeof(__VA_ARGS__)

typedef struct
{
    char *name;
    str format;
    fmt_args args;
    str buf;
    str wanted;
} fprintf_test;

static fprintf_test fprintf_tests[] = {
    {
        .name = "test_fmt_fprintf:empty",
        .format = STR(""),
        .args = FMT_ARGS(),
        .buf = STR_ARR((char[256]){0}),
        .wanted = STR(""),
    },
    {
        .name = "test_fmt_fprintf:no-directives",
        .format = STR("foo bar"),
        .args = FMT_ARGS(),
        .buf = STR_ARR((char[256]){0}),
        .wanted = STR("foo bar"),
    },
    {
        .name = "test_fmt_fprintf:one-directive",
        .format = STR("foo {} baz"),
        .args = FMT_ARGS(FMT_STR_LIT("bar")),
        .buf = STR_ARR((char[256]){0}),
        .wanted = STR("foo bar baz"),
    },
    {
        .name = "test_fmt_fprintf:one-directive-but-no-args",
        .format = STR("foo {} baz"),
        .args = FMT_ARGS(),
        .buf = STR_ARR((char[256]){0}),
        .wanted = STR("foo {}(MISSING) baz"),
    },
    {
        .name = "test_fmt_fprintf:multiple-arg-iterations",
        .format = STR("foo {} baz"),
        .args = FMT_ARGS(FMT_STR_LIT(
            "<this-string-is-longer-than-the-buffer>")),
        .buf = STR_ARR((char[3]){0}),
        .wanted = STR("foo <this-string-is-longer-than-the-buffer> baz"),
    },
    {
        .name = "test_fmt_fprintf:multiple-directives",
        .format = STR("foo {} baz {}"),
        .args = FMT_ARGS(FMT_STR_LIT("bar"), FMT_STR_LIT("qux")),
        .buf = STR_ARR((char[3]){0}),
        .wanted = STR("foo bar baz qux"),
    },
};

bool fprintf_test_run(fprintf_test *tc)
{
    test_init(tc->name);

    string found = string_new();
    TEST_DEFER(string_drop, &found);
    io_result res = fmt_fprintf_buf(
        string_writer(&found),
        tc->format,
        tc->args,
        tc->buf);

    if (io_result_is_err(res))
    {
        formatter f;
        string s = string_new();
        TEST_DEFER(string_drop, &s);
        string_formatter(&f, &s);
        if (!error_display(res.err, f))
        {
            return test_fail("failed to display error");
        }
        char msg[256] = {0};
        string_copy_to_c(msg, &s, sizeof(msg));
        return test_fail("unexpected err: %s", msg);
    }

    if (!str_eq(string_borrow(&found), tc->wanted))
    {
        char w[256] = {0}, f[256] = {0};
        str_copy_to_c(w, tc->wanted, sizeof(w));
        str_copy_to_c(f, string_borrow(&found), sizeof(f));
        return test_fail("wanted `%s`; found `%s`", w, f);
    }
    return test_success();
}

bool test_fmt_fprintf()
{
    for (size_t i = 0; i < sizeof(fprintf_tests) / sizeof(fprintf_test); i++)
    {
        if (!fprintf_test_run(&fprintf_tests[i]))
        {
            return false;
        }
    }

    return true;
}

bool fmt_tests()
{
    return test_fmt_fprintf();
}
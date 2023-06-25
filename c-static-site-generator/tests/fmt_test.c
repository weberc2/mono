#include "core/fmt/fmt_reader.h"
#include "core/fmt/fmt_fprintf.h"
#include "std/string/string.h"
#include "std/string/string_writer.h"
#include "test.h"

#define ARR(...) __VA_ARGS__, sizeof(__VA_ARGS__)

typedef struct
{
    char *name;
    str format;
    fmt_args args;
    str buf;
    str wanted;
} read_test;

static read_test read_tests[] = {
    // {
    //     .name = "test_fmt_fprintf:empty",
    //     .format = STR_LIT(""),
    //     .args = FMT_ARGS(),
    //     .buf = STR_ARR((char[256]){0}),
    //     .wanted = STR_LIT(""),
    // },
    // {
    //     .name = "test_fmt_fprintf:no-directives",
    //     .format = STR_LIT("foo bar"),
    //     .args = FMT_ARGS(),
    //     .buf = STR_ARR((char[256]){0}),
    //     .wanted = STR_LIT("foo bar"),
    // },
    // {
    //     .name = "test_fmt_fprintf:one-directive",
    //     .format = STR_LIT("foo {} baz"),
    //     .args = FMT_ARGS(FMT_STR_LIT("bar")),
    //     .buf = STR_ARR((char[256]){0}),
    //     .wanted = STR_LIT("foo bar baz"),
    // },
    // {
    //     .name = "test_fmt_fprintf:one-directive-but-no-args",
    //     .format = STR_LIT("foo {} baz"),
    //     .args = FMT_ARGS(),
    //     .buf = STR_ARR((char[256]){0}),
    //     .wanted = STR_LIT("foo {}(MISSING) baz"),
    // },
    {
        .name = "test_fmt_fprintf:multiple-arg-iterations",
        .format = STR_LIT("foo {} baz"),
        .args = FMT_ARGS(FMT_STR_LIT(
            "<this-string-is-longer-than-the-buffer>")),
        .buf = STR_ARR((char[3]){0}),
        .wanted = STR_LIT("foo <this-string-is-longer-than-the-buffer> baz"),
    },
    // {
    //     .name = "test_fmt_fprintf:multiple-directives",
    //     .format = STR_LIT("foo {} baz {}"),
    //     .args = FMT_ARGS(FMT_STR_LIT("bar"), FMT_STR_LIT("qux")),
    //     .buf = STR_ARR((char[3]){0}),
    //     .wanted = STR_LIT("foo bar baz qux"),
    // },
};

bool read_test_run(read_test *tc)
{
    test_init(tc->name);

    string found = string_new();
    TEST_DEFER(string_drop, &found);
    fmt_result res = fmt_fprintf_buf(
        string_writer(&found),
        tc->format,
        tc->args,
        tc->buf);

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
    for (size_t i = 0; i < sizeof(read_tests) / sizeof(read_test); i++)
    {
        if (!read_test_run(&read_tests[i]))
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
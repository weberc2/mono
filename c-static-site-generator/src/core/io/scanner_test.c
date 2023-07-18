#include <stdio.h>
#include <string.h>

#include "core/error/error.h"
#include "core/io/scanner.h"
#include "core/io/str_reader.h"
#include "core/fmt/str_formatter.h"
#include "core/testing/test.h"
#include "std/string/string.h"
#include "std/string/string_writer.h"

typedef struct scanner_write_to_test
{
    char *name;
    reader source;
    str buffer;
    str delim;
    bool wanted_new_ok;
    str wanted;
    error wanted_err;
} scanner_write_to_test;

#define TEST_READER(input) READER( \
    &STR_READER(STR(input)),       \
    str_reader_io_read)

#define BUFFER(size, init) STR_ARR((char[size]){(init)})

#define TEST_SCANNER(input, buf, del) SCANNER( \
    TEST_READER(input),                        \
    (buf),                                     \
    STR(del))

scanner_write_to_test scanner_write_to_tests[] = {
    // {
    //     .name = "empty",
    //     .source = TEST_READER(""),
    //     .buffer = BUFFER(16, 0),
    //     .delim = STR("---"),
    //     .wanted_new_ok = true,
    //     .wanted = STR(""),
    //     .wanted_err = ERROR_NULL,
    // },
    // {
    //     .name = "no-delim",
    //     .source = TEST_READER("foobar"),
    //     .buffer = BUFFER(3, 0),
    //     .delim = STR("---"),
    //     .wanted_new_ok = true,
    //     .wanted = STR("foobar"),
    //     .wanted_err = ERROR_NULL,
    // },
    // {
    //     .name = "delim-not-straddling-buffer-boundary",
    //     .source = TEST_READER("foo```bar"),
    //     .buffer = BUFFER(6, 0),
    //     .delim = STR("```"),
    //     .wanted_new_ok = true,
    //     .wanted = STR("foo"),
    //     .wanted_err = ERROR_NULL,
    // },
    // {
    //     .name = "delim-straddles-buffer-boundary",
    //     .source = TEST_READER("foo```bar"),
    //     .buffer = BUFFER(4, 0),
    //     .delim = STR("```"),
    //     .wanted_new_ok = true,
    //     .wanted = STR("foo"),
    //     .wanted_err = ERROR_NULL,
    // },
    // {
    //     .name = "first-iteration-ends-in-prefix-but-second-fails-to-match",
    //     .source = TEST_READER("foo--baz"),
    //     .buffer = BUFFER(4, 0),
    //     .delim = STR("---"),
    //     .wanted_new_ok = true,
    //     .wanted = STR("foo--baz"),
    //     .wanted_err = ERROR_NULL,
    // },
    // {
    //     .name = "back-to-back-partial-prefix-matches",
    //     .source = TEST_READER("foobabaz"),
    //     .buffer = BUFFER(4, 0),
    //     .delim = STR("bar"),
    //     .wanted_new_ok = true,
    //     .wanted = STR("foobabaz"),
    //     .wanted_err = ERROR_NULL,
    // },
    {
        .name = "final-iteration-ends-with-incomplete-prefix-then-eof",
        .source = TEST_READER("fooba"),
        .buffer = BUFFER(3, 0),
        .delim = STR("bar"),
        .wanted_new_ok = true,
        .wanted = STR("fooba"),
        .wanted_err = ERROR_NULL,
    },
    {
        .name = "delim-larger-than-buffer-is-error",
        .source = TEST_READER(""),
        .buffer = BUFFER(2, 0),
        .delim = STR("bar"),
        .wanted_new_ok = false,
        .wanted = STR(""),
        .wanted_err = ERROR_NULL,
    },

    // TODO: validate next_section()
    // TODO: validate io error handling
};

bool scanner_write_to_test_run(scanner_write_to_test *tc)
{
    char name[256] = {0};
    sprintf(name, "test_scanner_write_to:%s", tc->name);
    test_init(name);
    string found = string_new();
    TEST_DEFER(string_drop, &found);
    scanner_new_result new_res = scanner_new(
        tc->source,
        tc->buffer,
        tc->delim);
    if (new_res.ok != tc->wanted_new_ok)
    {
        return test_fail(
            "scanner_new_result.ok: wanted `%s`; found `%s`",
            tc->wanted_new_ok ? "true" : "false",
            new_res.ok ? "true" : "false");
    }

    // if the test was never meant to advance beyond scanner_new() then return
    // successful
    if (!tc->wanted_new_ok)
    {
        return test_success();
    }

    io_result res = scanner_write_to(&new_res.scanner, string_writer(&found));

    char *w = error_to_raw(tc->wanted_err, STR_BUF(256, 0));
    char *f = error_to_raw(res.err, STR_BUF(256, 0));
    if (strcmp(w, f) != 0)
    {
        return test_fail("err: wanted `%s`; found `%s`", w, f);
    }

    if (!str_eq(tc->wanted, string_borrow(&found)))
    {
        char w[256] = {0}, f[256] = {0};
        str_copy_to_c(w, tc->wanted, sizeof(w));
        string_copy_to_c(f, &found, sizeof(f));
        return test_fail("data: wanted `%s`; found `%s`", w, f);
    }

    return test_success();
}

bool test_scanner_write_to()
{
    for (
        size_t i = 0;
        i < sizeof(scanner_write_to_tests) / sizeof(scanner_write_to_test);
        i++)
    {
        if (!scanner_write_to_test_run(&scanner_write_to_tests[i]))
        {
            return false;
        }
    }
    return true;
}
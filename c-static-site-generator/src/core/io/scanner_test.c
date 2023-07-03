#include <stdio.h>

#include "core/error/error.h"
#include "core/io/scanner.h"
#include "core/io/str_reader.h"
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
    bool wanted_err;
} scanner_write_to_test;

#define READER(ptr, fn)          \
    (reader)                     \
    {                            \
        .data = (void *)(ptr),   \
        .read = (read_func)(fn), \
    }

#define TEST_READER(input) READER( \
    &STR_READER(STR_LIT(input)),   \
    str_reader_io_read)

#define BUFFER(size, init) STR_ARR((char[size]){(init)})

#define TEST_SCANNER(input, buf, del) SCANNER( \
    TEST_READER(input),                        \
    (buf),                                     \
    STR_LIT(del))

scanner_write_to_test scanner_write_to_tests[] = {
    {
        .name = "empty",
        .source = TEST_READER(""),
        .buffer = BUFFER(16, 0),
        .delim = STR_LIT("---"),
        .wanted_new_ok = true,
        .wanted = STR_LIT(""),
        .wanted_err = false,
    },
    {
        .name = "no-delim",
        .source = TEST_READER("foobar"),
        .buffer = BUFFER(3, 0),
        .delim = STR_LIT("---"),
        .wanted_new_ok = true,
        .wanted = STR_LIT("foobar"),
        .wanted_err = false,
    },
    {
        .name = "delim-not-straddling-buffer-boundary",
        .source = TEST_READER("foo```bar"),
        .buffer = BUFFER(6, 0),
        .delim = STR_LIT("```"),
        .wanted_new_ok = true,
        .wanted = STR_LIT("foo"),
        .wanted_err = false,
    },
    {
        .name = "delim-straddles-buffer-boundary",
        .source = TEST_READER("foo```bar"),
        .buffer = BUFFER(4, 0),
        .delim = STR_LIT("```"),
        .wanted_new_ok = true,
        .wanted = STR_LIT("foo"),
        .wanted_err = false,
    },
    {
        .name = "first-iteration-ends-in-prefix-but-second-fails-to-match",
        .source = TEST_READER("foo--baz"),
        .buffer = BUFFER(4, 0),
        .delim = STR_LIT("---"),
        .wanted_new_ok = true,
        .wanted = STR_LIT("foo--baz"),
        .wanted_err = false,
    },
    {
        .name = "back-to-back-partial-prefix-matches",
        .source = TEST_READER("foobabaz"),
        .buffer = BUFFER(4, 0),
        .delim = STR_LIT("bar"),
        .wanted_new_ok = true,
        .wanted = STR_LIT("foobabaz"),
        .wanted_err = false,
    },
    {
        .name = "final-iteration-ends-with-incomplete-prefix-then-eof",
        .source = TEST_READER("fooba"),
        .buffer = BUFFER(3, 0),
        .delim = STR_LIT("bar"),
        .wanted_new_ok = true,
        .wanted = STR_LIT("fooba"),
        .wanted_err = false,
    },
    {
        .name = "delim-larger-than-buffer-is-error",
        .source = TEST_READER(""),
        .buffer = BUFFER(2, 0),
        .delim = STR_LIT("bar"),
        .wanted_new_ok = false,
        .wanted = STR_LIT(""),
        .wanted_err = false,
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
    result res = result_new();
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

    scanner_write_to(&new_res.scanner, string_writer(&found), &res);

    if (res.ok != tc->wanted_err)
    {
        return test_fail(
            "err: wanted `%s`; found `%s`",
            tc->wanted_err ? "true" : "false",
            res.ok ? "true" : "false");
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

int main()
{
    if (!test_scanner_write_to())
    {
        return 1;
    }
    return 0;
}
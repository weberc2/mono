#include <stdbool.h>
#include "core/str/str.h"
#include "test.h"

typedef struct
{
    char *name;
    str input;
    str prefix;
    bool wanted_match;
} has_prefix_test;

static has_prefix_test has_prefix_tests[] = {
    {
        .name = "str_has_prefix:single-char-match",
        .input = STR_FROM_CSTR("_hello"),
        .prefix = STR_FROM_CSTR("_"),
        .wanted_match = true,
    },
    {
        .name = "str_has_prefix:single-char-no-match",
        .input = STR_FROM_CSTR("_hello"),
        .prefix = STR_FROM_CSTR("-"),
        .wanted_match = false,
    },
    {
        .name = "str_has_prefix:long-prefix-match",
        .input = STR_FROM_CSTR("hello world"),
        .prefix = STR_FROM_CSTR("hello "),
        .wanted_match = true,
    },
    {
        .name = "str_has_prefix:long-prefix-no-match",
        .input = STR_FROM_CSTR("hello world"),
        .prefix = STR_FROM_CSTR("hello\t"),
        .wanted_match = false,
    },
};

bool has_prefix_test_run(has_prefix_test *tc)
{
    test_init(tc->name);
    bool found_match = str_has_prefix(tc->input, tc->prefix);
    if (!found_match && tc->wanted_match)
    {
        return test_fail("wanted match but found none");
    }
    if (found_match && !tc->wanted_match)
    {
        return test_fail("wanted no match but found a match");
    }
    return test_success();
}

bool test_str_has_prefix()
{
    for (size_t i = 0; i < sizeof(has_prefix_tests) / sizeof(has_prefix_test); i++)
    {
        if (!has_prefix_test_run(&has_prefix_tests[i]))
        {
            return false;
        }
    }

    return true;
}

typedef struct
{
    char *name;
    str (*trim)(str, str);
    str input;
    str cutset;
    str wanted;
} trim_test;

static trim_test trim_tests[] = {
    {
        .name = "str_trim_left:single-char-cutset-match",
        .trim = str_trim_left,
        .input = STR_FROM_CSTR("_hello_"),
        .cutset = STR_FROM_CSTR("_"),
        .wanted = STR_FROM_CSTR("hello_"),
    },
    {
        .name = "str_trim_left:multi-char-cutset-single-match-first-char",
        .trim = str_trim_left,
        .input = STR_FROM_CSTR("_hello_"),
        .cutset = STR_FROM_CSTR("_-"),
        .wanted = STR_FROM_CSTR("hello_"),
    },
    {
        .name = "str_trim_left:multi-char-cutset-single-match-second-char",
        .trim = str_trim_left,
        .input = STR_FROM_CSTR("_hello_"),
        .cutset = STR_FROM_CSTR("-_"),
        .wanted = STR_FROM_CSTR("hello_"),
    },
    {
        .name = "str_trim_left:multi-char-cutset-single-match-multi-char",
        .trim = str_trim_left,
        .input = STR_FROM_CSTR("_-hello-_"),
        .cutset = STR_FROM_CSTR("-_"),
        .wanted = STR_FROM_CSTR("hello-_"),
    },
    {
        .name = "str_trim_left:multi-char-cutset-no-match",
        .trim = str_trim_left,
        .input = STR_FROM_CSTR("_hello_"),
        .cutset = STR_FROM_CSTR("!@"),
        .wanted = STR_FROM_CSTR("_hello_"),
    },
    {
        .name = "str_trim_right:single-char-cutset-match",
        .trim = str_trim_right,
        .input = STR_FROM_CSTR("_hello_"),
        .cutset = STR_FROM_CSTR("_"),
        .wanted = STR_FROM_CSTR("_hello"),
    },
    {
        .name = "str_trim_right:multi-char-cutset-single-match-first-char",
        .trim = str_trim_right,
        .input = STR_FROM_CSTR("_hello_"),
        .cutset = STR_FROM_CSTR("_-"),
        .wanted = STR_FROM_CSTR("_hello"),
    },
    {
        .name = "str_trim_right:multi-char-cutset-single-match-second-char",
        .trim = str_trim_right,
        .input = STR_FROM_CSTR("_hello_"),
        .cutset = STR_FROM_CSTR("-_"),
        .wanted = STR_FROM_CSTR("_hello"),
    },
    {
        .name = "str_trim_right:multi-char-cutset-single-match-multi-char",
        .trim = str_trim_right,
        .input = STR_FROM_CSTR("_-hello-_"),
        .cutset = STR_FROM_CSTR("-_"),
        .wanted = STR_FROM_CSTR("_-hello"),
    },
    {
        .name = "str_trim_right:multi-char-cutset-no-match",
        .trim = str_trim_right,
        .input = STR_FROM_CSTR("_hello_"),
        .cutset = STR_FROM_CSTR("!@"),
        .wanted = STR_FROM_CSTR("_hello_"),
    },
    {
        .name = "str_trim:single-char-cutset-match",
        .trim = str_trim,
        .input = STR_FROM_CSTR("_hello_"),
        .cutset = STR_FROM_CSTR("_"),
        .wanted = STR_FROM_CSTR("hello"),
    },
    {
        .name = "str_trim:multi-char-cutset-single-match-first-char",
        .trim = str_trim,
        .input = STR_FROM_CSTR("_hello_"),
        .cutset = STR_FROM_CSTR("_-"),
        .wanted = STR_FROM_CSTR("hello"),
    },
    {
        .name = "str_trim:multi-char-cutset-single-match-second-char",
        .trim = str_trim,
        .input = STR_FROM_CSTR("_hello_"),
        .cutset = STR_FROM_CSTR("-_"),
        .wanted = STR_FROM_CSTR("hello"),
    },
    {
        .name = "str_trim:multi-char-cutset-single-match-multi-char",
        .trim = str_trim,
        .input = STR_FROM_CSTR("_-hello-_"),
        .cutset = STR_FROM_CSTR("-_"),
        .wanted = STR_FROM_CSTR("hello"),
    },
    {
        .name = "str_trim:multi-char-cutset-no-match",
        .trim = str_trim,
        .input = STR_FROM_CSTR("_hello_"),
        .cutset = STR_FROM_CSTR("!@"),
        .wanted = STR_FROM_CSTR("_hello_"),
    },
};

bool trim_test_run(trim_test *tc)
{
    test_init(tc->name);
    str found = tc->trim(tc->input, tc->cutset);
    if (!str_eq(found, tc->wanted))
    {
        char f[256] = {0}, w[256] = {0};
        str_copy_to_c(f, found, sizeof(f));
        str_copy_to_c(w, tc->wanted, sizeof(w));
        return test_fail("wanted `%s`; found `%s`", w, f);
    }
    return test_success();
}

bool test_trim()
{
    for (size_t i = 0; i < sizeof(trim_tests) / sizeof(trim_test); i++)
    {
        if (!trim_test_run(&trim_tests[i]))
        {
            return false;
        }
    }
    return true;
}

bool str_tests()
{
    return test_str_has_prefix() && test_trim();
}
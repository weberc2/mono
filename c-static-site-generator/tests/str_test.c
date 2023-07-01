#include <stdbool.h>
#include "core/str/str.h"
#include "core/testing/test.h"

typedef struct
{
    char *name;
    str input;
    str prefix;
    bool wanted_match;
} has_prefix_test;

static has_prefix_test has_prefix_tests[] = {
    {
        .name = "test_str_has_prefix:single-char-match",
        .input = STR_LIT("_hello"),
        .prefix = STR_LIT("_"),
        .wanted_match = true,
    },
    {
        .name = "test_str_has_prefix:single-char-no-match",
        .input = STR_LIT("_hello"),
        .prefix = STR_LIT("-"),
        .wanted_match = false,
    },
    {
        .name = "test_str_has_prefix:long-prefix-match",
        .input = STR_LIT("hello world"),
        .prefix = STR_LIT("hello "),
        .wanted_match = true,
    },
    {
        .name = "test_str_has_prefix:long-prefix-no-match",
        .input = STR_LIT("hello world"),
        .prefix = STR_LIT("hello\t"),
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
        .name = "test_str_trim_left:single-char-cutset-match",
        .trim = str_trim_left,
        .input = STR_LIT("_hello_"),
        .cutset = STR_LIT("_"),
        .wanted = STR_LIT("hello_"),
    },
    {
        .name = "test_str_trim_left:multi-char-cutset-single-match-first-char",
        .trim = str_trim_left,
        .input = STR_LIT("_hello_"),
        .cutset = STR_LIT("_-"),
        .wanted = STR_LIT("hello_"),
    },
    {
        .name = "test_str_trim_left:multi-char-cutset-single-match-second-char",
        .trim = str_trim_left,
        .input = STR_LIT("_hello_"),
        .cutset = STR_LIT("-_"),
        .wanted = STR_LIT("hello_"),
    },
    {
        .name = "test_str_trim_left:multi-char-cutset-single-match-multi-char",
        .trim = str_trim_left,
        .input = STR_LIT("_-hello-_"),
        .cutset = STR_LIT("-_"),
        .wanted = STR_LIT("hello-_"),
    },
    {
        .name = "test_str_trim_left:multi-char-cutset-no-match",
        .trim = str_trim_left,
        .input = STR_LIT("_hello_"),
        .cutset = STR_LIT("!@"),
        .wanted = STR_LIT("_hello_"),
    },
    {
        .name = "test_str_trim_right:single-char-cutset-match",
        .trim = str_trim_right,
        .input = STR_LIT("_hello_"),
        .cutset = STR_LIT("_"),
        .wanted = STR_LIT("_hello"),
    },
    {
        .name = "test_str_trim_right:multi-char-cutset-single-match-first-char",
        .trim = str_trim_right,
        .input = STR_LIT("_hello_"),
        .cutset = STR_LIT("_-"),
        .wanted = STR_LIT("_hello"),
    },
    {
        .name = "test_str_trim_right:multi-char-cutset-single-match-second-char",
        .trim = str_trim_right,
        .input = STR_LIT("_hello_"),
        .cutset = STR_LIT("-_"),
        .wanted = STR_LIT("_hello"),
    },
    {
        .name = "test_str_trim_right:multi-char-cutset-single-match-multi-char",
        .trim = str_trim_right,
        .input = STR_LIT("_-hello-_"),
        .cutset = STR_LIT("-_"),
        .wanted = STR_LIT("_-hello"),
    },
    {
        .name = "test_str_trim_right:multi-char-cutset-no-match",
        .trim = str_trim_right,
        .input = STR_LIT("_hello_"),
        .cutset = STR_LIT("!@"),
        .wanted = STR_LIT("_hello_"),
    },
    {
        .name = "test_str_trim:single-char-cutset-match",
        .trim = str_trim,
        .input = STR_LIT("_hello_"),
        .cutset = STR_LIT("_"),
        .wanted = STR_LIT("hello"),
    },
    {
        .name = "test_str_trim:multi-char-cutset-single-match-first-char",
        .trim = str_trim,
        .input = STR_LIT("_hello_"),
        .cutset = STR_LIT("_-"),
        .wanted = STR_LIT("hello"),
    },
    {
        .name = "test_str_trim:multi-char-cutset-single-match-second-char",
        .trim = str_trim,
        .input = STR_LIT("_hello_"),
        .cutset = STR_LIT("-_"),
        .wanted = STR_LIT("hello"),
    },
    {
        .name = "test_str_trim:multi-char-cutset-single-match-multi-char",
        .trim = str_trim,
        .input = STR_LIT("_-hello-_"),
        .cutset = STR_LIT("-_"),
        .wanted = STR_LIT("hello"),
    },
    {
        .name = "test_str_trim:multi-char-cutset-no-match",
        .trim = str_trim,
        .input = STR_LIT("_hello_"),
        .cutset = STR_LIT("!@"),
        .wanted = STR_LIT("_hello_"),
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

typedef struct
{
    char *name;
    str input;
    str match;
    str_find_result wanted;
} find_test;

static find_test find_tests[] = {
    {
        .name = "test_str_find:prefix-match",
        .input = STR_LIT("foo bar baz"),
        .match = STR_LIT("foo"),
        .wanted = {.found = true, .index = 0},
    },
    {
        .name = "test_str_find:mid-match",
        .input = STR_LIT("foo bar baz"),
        .match = STR_LIT("bar"),
        .wanted = {.found = true, .index = 4},
    },
    {
        .name = "test_str_find:suffix-match",
        .input = STR_LIT("foo bar baz"),
        .match = STR_LIT("baz"),
        .wanted = {.found = true, .index = 8},
    },
    {
        .name = "test_str_find:no-match",
        .input = STR_LIT("foo bar baz"),
        .match = STR_LIT("qux"),
        .wanted = {.found = false, .index = 0},
    },
    {
        .name = "test_str_find:match-longer-than-input",
        .input = STR_LIT("foo"),
        .match = STR_LIT("foobar"),
        .wanted = {.found = false, .index = 0},
    },
};

static bool find_test_run(find_test *tc)
{
    test_init(tc->name);
    str_find_result found = str_find(tc->input, tc->match);
    if (found.found && !tc->wanted.found)
    {
        char m[256] = {0}, i[256] = {0};
        str_copy_to_c(m, tc->match, sizeof(m));
        str_copy_to_c(i, tc->input, sizeof(i));
        return test_fail("unexpectedly found `%s` in `%s`", m, i);
    }
    if (!found.found && tc->wanted.found)
    {
        char m[256] = {0}, i[256] = {0};
        str_copy_to_c(m, tc->match, sizeof(m));
        str_copy_to_c(i, tc->input, sizeof(i));
        return test_fail("expected to find `%s` in `%s` but failed", m, i);
    }
    if (found.index != tc->wanted.index)
    {
        char m[256] = {0}, i[256] = {0};
        str_copy_to_c(m, tc->match, sizeof(m));
        str_copy_to_c(i, tc->input, sizeof(i));
        return test_fail(
            "str_find(`%s`, `%s`).index: wanted `%zu`; found `%zu`",
            i,
            m,
            tc->wanted.index,
            found.index);
    }

    return test_success();
}

static bool test_str_find()
{
    for (size_t i = 0; i < sizeof(find_tests) / sizeof(find_test); i++)
    {
        if (!find_test_run(&find_tests[i]))
        {
            return false;
        }
    }

    return true;
}

typedef struct
{
    char *name;
    str input;
    char match;
    str_find_result wanted;
} find_char_test;

static find_char_test find_char_tests[] = {
    {
        .name = "test_str_find_char:prefix-match",
        .input = STR_LIT("abc"),
        .match = 'a',
        .wanted = {.found = true, .index = 0},
    },
    {
        .name = "test_str_find_char:mid-match",
        .input = STR_LIT("abc"),
        .match = 'b',
        .wanted = {.found = true, .index = 1},
    },
    {
        .name = "test_str_find_char:suffix-match",
        .input = STR_LIT("abc"),
        .match = 'c',
        .wanted = {.found = true, .index = 2},
    },
    {
        .name = "test_str_find_char:no-match",
        .input = STR_LIT("abc"),
        .match = 'z',
        .wanted = {.found = false, .index = 0},
    },
};

static bool find_char_test_run(find_char_test *tc)
{
    test_init(tc->name);
    str_find_result found = str_find_char(tc->input, tc->match);
    if (found.found && !tc->wanted.found)
    {
        char i[256] = {0};
        str_copy_to_c(i, tc->input, sizeof(i));
        return test_fail("unexpectedly found `%c` in `%s`", tc->match, i);
    }
    if (!found.found && tc->wanted.found)
    {
        char i[256] = {0};
        str_copy_to_c(i, tc->input, sizeof(i));
        return test_fail(
            "expected to find_char `%c` in `%s` but failed",
            tc->match,
            i);
    }
    if (found.index != tc->wanted.index)
    {
        char i[256] = {0};
        str_copy_to_c(i, tc->input, sizeof(i));
        return test_fail(
            "str_find_char(`%s`, `%c`).index: wanted `%zu`; found `%zu`",
            i,
            tc->match,
            tc->wanted.index,
            found.index);
    }

    return test_success();
}

static bool test_str_find_char()
{
    for (size_t i = 0; i < sizeof(find_char_tests) / sizeof(find_char_test); i++)
    {
        if (!find_char_test_run(&find_char_tests[i]))
        {
            return false;
        }
    }

    return true;
}

bool str_tests()
{
    return test_str_has_prefix() &&
           test_trim() &&
           test_str_find() &&
           test_str_find_char();
}
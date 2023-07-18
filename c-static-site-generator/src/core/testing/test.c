
#include <stdarg.h>
#include <stdio.h>
#include <stdbool.h>
#include "core/testing/test.h"
#include "core/panic/panic.h"
#include "std/vector/vector.h"
#include "std/string/string.h"
#include "std/string/string_formatter.h"

char *current_test;

struct deferable
{
    void *data;
    defer_func defer;
};

static vector deferables;
static bool initialized;

void test_init(char *name)
{
    current_test = name;
    if (!initialized)
    {
        vector_init(&deferables, sizeof(struct deferable));
        initialized = true;
    }
}

void test_defer(defer_func func, void *data)
{
    struct deferable deferable = {data, func};
    vector_push(&deferables, &deferable);
}

void test_run_defer()
{
    struct deferable deferable;
    while (vector_pop(&deferables, &deferable))
    {
        deferable.defer(deferable.data);
    }
}

bool test_fail(const char *format, ...)
{
    test_run_defer();

    printf("FAIL: %s(): ", current_test);
    va_list args;
    va_start(args, format);
    vprintf(format, args);
    va_end(args);
    printf("\n");
    return false;
}

bool test_success()
{
    test_run_defer();

    printf("SUCCESS: %s()\n", current_test);
    return true;
}

char *error_to_raw(error err, str buf)
{
    string s = string_new();
    formatter f;
    string_formatter(&f, &s);

    if (!error_display(err, f))
    {
        panic("failed to display error!");
    }

    str_copy(buf, string_borrow(&s));
    string_drop(&s);
    return buf.data;
}

bool assert_ok(io_result res)
{
    if (io_result_is_ok(res))
    {
        return true;
    }

    return test_fail(
        "unexpected err: %s",
        error_to_raw(res.err, STR_ARR((char[256]){0})));
}
#include "test.h"

#include <stdarg.h>
#include <stdio.h>
#include <stdbool.h>
#include "vector/vector.h"
#include "string/string.h"
#include "string/string_formatter.h"

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

char *error_to_raw(error err)
{
    string s;
    string_init(&s);
    TEST_DEFER(string_drop, &s);

    formatter f;
    string_formatter(&f, &s);

    error_display(err, f);
    char *tmp = calloc(s.len, 1);
    TEST_DEFER(free, tmp);

    string_copy_to_c(tmp, &s, s.len);
    return tmp;
}

bool assert_ok(io_result res)
{
    if (res.ok)
    {
        return true;
    }

    return test_fail("unexpected err: %s", error_to_raw(res.err));
}
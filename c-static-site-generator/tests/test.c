#include "test.h"

#include <stdarg.h>
#include <stdio.h>
#include <stdbool.h>
#include "vector/vector.h"

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
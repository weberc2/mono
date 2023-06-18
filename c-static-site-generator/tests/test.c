#include "test.h"

#include <stdarg.h>
#include <stdio.h>
#include <stdbool.h>

char *current_test;

void test_init(char *name)
{
    current_test = name;
}

static void noop(void *_) {}

static defer_func deferred = noop;
void *deferred_data;

void test_defer(defer_func func, void *data)
{
    deferred = func;
    deferred_data = data;
}

void test_run_defer()
{
    deferred(deferred_data);
    deferred = noop;
    deferred_data = NULL;
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

void defer_many(vector *deferables)
{
    for (int i = deferables->len - 1; i >= 0; i--)
    {
        struct deferable *deferable = vector_get(deferables, i);
        deferable->defer(deferable->data);
    }
}

void deferables_push(vector *deferables, void *data, defer_func defer)
{
    struct deferable deferable = {data, defer};
    vector_push(deferables, &deferable);
}
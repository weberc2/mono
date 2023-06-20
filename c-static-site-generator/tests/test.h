#ifndef TEST_H
#define TEST_H

#include <stdbool.h>
#include "io/io_result.h"

typedef void (*defer_func)(void *);

#define TEST_DEFER(fn, data) test_defer((defer_func)(fn), (data));

void test_init(char *name);
void test_defer(defer_func fn, void *data);
bool test_fail(const char *format, ...);
bool test_success();
bool assert_ok(io_result res);

#define ASSERT_OK(res)     \
    if (!assert_ok((res))) \
    {                      \
        return false;      \
    }

#endif // TEST_H
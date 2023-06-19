#ifndef TEST_H
#define TEST_H

#include <stdbool.h>

typedef void (*defer_func)(void *);

#define TEST_DEFER(fn, data) test_defer((defer_func)(fn), (data));

void test_module_init();
void test_init(char *name);
void test_defer(defer_func fn, void *data);
void test_run_defer();
bool test_fail(const char *format, ...);
bool test_success();

#endif // TEST_H
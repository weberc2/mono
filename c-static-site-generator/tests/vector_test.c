#include "test.h"
#include "vector/vector.h"
#include <stdbool.h>

bool test_vector_init()
{
    test_init("test_vector_init");

    vector v;
    vector_init(&v, sizeof(size_t));
    TEST_DEFER(vector_drop, &v);

    if (v.len != 0)
    {
        return test_fail("len: wanted `0`; found `%zu`", v.len);
    }

    if (v.cap != 0)
    {
        return test_fail("cap: wanted `0`; found `%zu`", v.cap);
    }

    if (v.data != NULL)
    {
        return test_fail("data: wanted `NULL`; found `%p`", v.data);
    }

    return test_success();
}

bool test_vector_push__grow()
{
    test_init("test_vector_push__grow");

    vector v;
    vector_init(&v, sizeof(size_t));
    TEST_DEFER(vector_drop, &v);

    size_t value = 0;
    vector_push(&v, &value);

    if (v.len != 1)
    {
        return test_fail("len: wanted `1`; found `%zu`", v.len);
    }

    size_t found = *((size_t *)vector_get(&v, 0));
    if (found != value)
    {
        return test_fail("get(0): wanted `%zu`; found `%zu`", value, found);
    }

    return test_success();
}

bool test_vector_push__no_grow()
{
    test_init("test_vector_push__no_grow");

    vector v;
    vector_init_with_cap(&v, sizeof(size_t), 1);
    TEST_DEFER(vector_drop, &v);

    size_t value = 1;
    vector_push(&v, &value);

    if (v.len != 1)
    {
        return test_fail("len: wanted `1`; found `%zu`", v.len);
    }

    size_t found = *((size_t *)vector_get(&v, 0));
    if (found != value)
    {
        return test_fail("get(0): wanted `%zu`; found `%zu`", value, found);
    }

    return test_success();
}

bool vector_tests()
{
    return test_vector_init() &&
           test_vector_push__grow() &&
           test_vector_push__no_grow();
}
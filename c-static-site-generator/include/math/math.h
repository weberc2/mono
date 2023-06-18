#ifndef MATH_H
#define MATH_H

#include <stddef.h>

static inline size_t min(size_t a, size_t b)
{
    if (a < b)
    {
        return a;
    }
    return b;
}

static inline size_t max(size_t a, size_t b)
{
    if (a > b)
    {
        return a;
    }
    return b;
}

#endif // MATH_H
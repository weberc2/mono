#ifndef RESULT_H
#define RESULT_H

#include "core/error/error.h"

typedef struct
{
    bool ok;
    error err;
} result;

#define RESULT_OK         \
    (result)              \
    {                     \
        .ok = true,       \
        .err = ERROR_NULL \
    }

#define RESULT_ERR(e) \
    (result)          \
    {                 \
        .ok = false,  \
        .err = e,     \
    }

void result_init(result *res);
result result_new();
result result_ok();
result result_err(error err);

#endif // RESULT_H
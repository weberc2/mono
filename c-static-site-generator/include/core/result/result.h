#ifndef RESULT_H
#define RESULT_H

#include "core/error/error.h"

typedef struct
{
    bool ok;
    error err;
} result;

void result_init(result *res);
result result_new();
result result_ok();
result result_err(error err);

#endif // RESULT_H
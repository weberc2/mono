#ifndef RESULT_H
#define RESULT_H

#include "core/error/error.h"

typedef struct
{
    bool ok;
    error err;
} result;

void result_init(result *res);
void result_ok(result *res);
void result_err(result *res, error err);

#endif // RESULT_H
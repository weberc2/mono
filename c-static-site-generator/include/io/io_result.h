#ifndef IO_RESULT_H
#define IO_RESULT_H

#include "error/error.h"

typedef struct
{
    bool ok;
    error err;
} io_result;

void io_result_ok(io_result *res);
void io_result_err(io_result *res, error err);

#endif // IO_RESULT_H
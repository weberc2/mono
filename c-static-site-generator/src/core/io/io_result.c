#include <stdbool.h>
#include "core/io/io_result.h"

void io_result_ok(io_result *res)
{
    res->ok = true;
    error_null(&res->err);
}

void io_result_err(io_result *res, error err)
{
    res->ok = false;
    res->err = err;
}
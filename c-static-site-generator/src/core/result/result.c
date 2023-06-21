#include <stdbool.h>
#include "core/result/result.h"

void result_init(result *res)
{
    res->ok = false;
    error_const(&res->err, "program error: result not initialized");
}

void result_ok(result *res)
{
    res->ok = true;
    error_null(&res->err);
}

void result_err(result *res, error err)
{
    res->ok = false;
    res->err = err;
}
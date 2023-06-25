#include <stdbool.h>
#include "core/result/result.h"

void result_init(result *res)
{
    res->ok = false;
    error_const(&res->err, "program error: result not initialized");
}

result result_new()
{
    result res;
    result_init(&res);
    return res;
}

result result_ok()
{
    result res;
    res.ok = true;
    res.err = error_null();
    return res;
}

result result_err(error err)
{
    return (result){.ok = false, .err = err};
}
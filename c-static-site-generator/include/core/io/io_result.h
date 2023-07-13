#ifndef IO_RESULT_H
#define IO_RESULT_H

#include "core/error/error.h"

typedef struct io_result
{
    size_t size;
    error err;
} io_result;

#define IO_RESULT(sz, e) \
    (io_result) { .size = (sz), .err = (e) }
#define IO_RESULT_OK(sz) IO_RESULT(sz, ERROR_NULL)
#define IO_RESULT_ERR(e) IO_RESULT(0, e)

static inline bool io_result_is_ok(io_result res)
{
    return error_is_null(res.err);
}

static inline bool io_result_is_err(io_result res)
{
    return !error_is_null(res.err);
}

#endif // IO_RESULT_H
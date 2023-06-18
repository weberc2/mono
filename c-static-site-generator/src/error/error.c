#include "error/error.h"

void error_write_string(error err, string *message)
{
    err.error_func(err.data, message);
}

void error_const_write_string(char *message, string *out)
{
    string_push_raw(out, message, strlen(message));
}

void error_const(error *err, const char *message)
{
    err->data = (void *)message;
    err->error_func = (error_func)error_const_write_string;
}

void errors_init(errors *errs)
{
    vector_init(&errs->errors, sizeof(error));
}

void errors_drop(errors *errs)
{
    vector_drop(&errs->errors);
}

void errors_push(errors *errs, error err)
{
    vector_push(&errs->errors, &err);
}

void errors_write_string(errors *errs, string *message)
{
    for (int i = errs->errors.len; i >= 0; i--)
    {
        error err = *(error *)vector_get((vector *)errs, i);
        string_push_raw(message, ": ", 2);
        error_write_string(err, message);
    }
}

size_t errors_len(errors *errs)
{
    return errs->errors.len;
}
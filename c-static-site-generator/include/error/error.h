#ifndef ERROR_H
#define ERROR_H

#include "string/string.h"
#include "vector/vector.h"

typedef void (*error_func)(void *, string *);

typedef struct
{
    void *data;
    error_func error_func;
} error;

void error_write_string(error err, string *message);
void error_const(error *err, const char *message);

typedef struct
{
    vector errors;
} errors;

void errors_init(errors *errs);
void errors_drop(errors *errs);
void errors_push(errors *errs, error err);
void errors_write_string(errors *errs, string *message);
size_t errors_len(errors *errs);

#endif // ERROR_H
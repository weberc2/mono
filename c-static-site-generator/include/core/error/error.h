#ifndef ERROR_H
#define ERROR_H

#include "core/fmt/display.h"

typedef struct
{
    void *data;
    display_func display;
} error;

#define ERROR_NULL \
    (error) { .data = NULL, .display = NULL }

void error_init(error *err, void *data, display_func display);
error error_null();
bool error_display(error err, formatter f);
void error_const(error *err, const char *message);

static inline bool error_is_null(error err)
{
    return err.data == NULL && err.display == NULL;
}

#endif // ERROR_H
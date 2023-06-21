#ifndef ERROR_H
#define ERROR_H

#include "core/fmt/display.h"

typedef struct
{
    void *data;
    display_func display;
} error;

void error_init(error *err, void *data, display_func display);
void error_null(error *err);
bool error_display(error err, formatter f);
void error_const(error *err, const char *message);

#endif // ERROR_H
#include "core/error/error.h"
#include <string.h>

void error_init(error *err, void *data, display_func display)
{
    err->data = data;
    err->display = display;
}

void error_null(error *err)
{
    err->data = NULL;
    err->display = NULL;
}

bool error_display(error err, formatter f)
{
    return err.display(err.data, f);
}

bool error_const_display(const char *message, formatter f)
{
    return fmt_write_str(f, str_new((char *)message, strlen(message)));
}

void error_const(error *err, const char *message)
{
    err->data = (void *)message;
    err->display = (display_func)error_const_display;
}

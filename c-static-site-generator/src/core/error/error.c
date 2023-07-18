#include "core/error/error.h"
#include <string.h>

void error_init(error *err, void *data, display_func display)
{
    err->data = data;
    err->display = display;
}

error error_null()
{
    return (error){.data = NULL, .display = NULL};
}

bool error_display(error err, formatter f)
{
    if (error_is_null(err))
    {
        return fmt_write_str(f, STR("<NULL>"));
    }
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

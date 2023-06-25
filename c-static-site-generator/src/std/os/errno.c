#include <string.h>
#include <errno.h>
#include "std/os/errno.h"

typedef int errno_t;

bool errno_display(errno_t err, formatter f)
{
    char *msg = strerror((int)err);
    return fmt_write_str(f, str_new(msg, strlen(msg)));
}

error errno_error(int errno)
{
    error err;
    error_init(&err, (void *)(uintptr_t)(errno), (display_func)errno_display);
    return err;
}
#ifndef ERR_EOF_H
#define ERR_EOF_H

#include "core/error/error.h"
#include "core/fmt/str_formatter.h"

#define ERR_EOF                                   \
    (error)                                       \
    {                                             \
        .data = NULL,                             \
        .display = (display_func)err_eof_display, \
    }

#define EOF_STR STR("end of file")

static inline bool err_eof_display(void *ptr, formatter f)
{
    return fmt_write_str(f, EOF_STR);
}

static inline bool error_is_eof(error err)
{
    // we can't just check that `err.data == ERR_EOF.data && err.display ==
    // ERR_EOF.display` because ERR_EOF.display will always point to
    // `err_eof_display`; however, inlining means `err_eof_display` may not
    // always have the same address. instead, we compare the strings.
    str_formatter sf = STR_FORMATTER_WITH_CAP(256);
    return error_display(err, str_formatter_to_formatter(&sf)) &&
           str_eq(EOF_STR, str_formatter_data(&sf));
}

#undef EOF_STR

#endif // ERR_EOF_H
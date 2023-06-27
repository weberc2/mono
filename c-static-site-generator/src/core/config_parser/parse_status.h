#ifndef PARSE_STATUS_H
#define PARSE_STATUS_H

#include "core/str/str.h"
#include "core/panic/panic.h"

typedef enum parse_status
{
    parse_ok,
    parse_io_error,
    parse_match_failure,
} parse_status;

static inline str parse_status_str(parse_status status)
{
    switch (status)
    {
    case parse_ok:
        return STR_LIT("PARSE_STATUS_OK");
    case parse_io_error:
        return STR_LIT("PARSE_STATUS_IO_ERROR");
    case parse_match_failure:
        return STR_LIT("PARSE_STATUS_MATCH_FAILURE");
    default:
        panic("invalid parse status: `%d`", status);
    }
}

#endif // PARSE_STATUS_H
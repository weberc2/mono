#ifndef FORMATTER_H
#define FORMATTER_H

#include <stdbool.h>
#include "core/str/str.h"

typedef bool (*formatter_write_str)(void *, str);

typedef struct
{
    void *data;
    formatter_write_str write_str;
} formatter;

bool fmt_write_str(formatter f, str s);

#endif // FORMATTER_H
#include <stdbool.h>
#include "string/string_formatter.h"
#include "string/string.h"
#include "core/fmt/formatter.h"

bool string_write_str(string *s, str str)
{
    string_push_slice(s, str);
    return true;
}

void string_formatter(formatter *f, string *s)
{
    f->data = (void *)s;
    f->write_str = (formatter_write_str)string_write_str;
}
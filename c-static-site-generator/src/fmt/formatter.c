#include <stdbool.h>
#include "fmt/formatter.h"

bool fmt_write_str(formatter f, str s)
{
    return f.write_str(f.data, s);
}
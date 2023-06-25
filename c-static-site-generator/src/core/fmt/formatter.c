#include <stdbool.h>
#include "core/panic/panic.h"
#include "core/fmt/formatter.h"
#include "core/io/writer.h"

bool fmt_write_str(formatter f, str s)
{
    return f.write_str(f.data, s);
}
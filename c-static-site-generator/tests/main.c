#include <stdbool.h>

bool str_tests();
bool vector_tests();
bool io_tests();
bool buffered_reader_find_tests();
bool fmt_tests();

#include "std/os/file.h"
#include "core/fmt/fmt_fprintf.h"

int main()
{
    if (
        str_tests() &&
        vector_tests() &&
        io_tests() &&
        buffered_reader_find_tests() &&
        fmt_tests())
    {
        return 0;
    }
    return 1;
}
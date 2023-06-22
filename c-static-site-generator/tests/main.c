#include <stdbool.h>

bool vector_tests();
bool io_tests();
bool buffered_reader_find_tests();

int main()
{
    if (vector_tests() && io_tests() && buffered_reader_find_tests())
    {
        return 0;
    }
    return 1;
}
#include <stdbool.h>

bool vector_tests();
bool io_tests();

int main()
{
    if (
        vector_tests() &&
        io_tests())
    {
        return 0;
    }
    return 1;
}
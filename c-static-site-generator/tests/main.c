#include <stdbool.h>

bool str_tests();
bool vector_tests();
bool fmt_tests();

int main()
{
    if (
        str_tests() &&
        vector_tests() &&
        fmt_tests())
    {
        return 0;
    }
    return 1;
}
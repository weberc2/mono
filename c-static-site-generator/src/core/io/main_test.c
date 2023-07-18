#include <stdbool.h>

extern bool test_str_reader();
extern bool test_copy();
extern bool test_scanner_write_to();

int main()
{
    return test_str_reader() && test_copy() && test_scanner_write_to() ? 0 : 1;
}
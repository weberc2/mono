#include <stdbool.h>

extern bool test_fields_match_name();
extern bool test_parse_field_name();
extern bool test_parse_field_value();

int main()
{
    if (test_fields_match_name() &&
        test_parse_field_name() &&
        test_parse_field_value())
    {
        return 0;
    }
    return 1;
}
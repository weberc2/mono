#include "core/panic/panic.h"
#include "core/io/str_reader.h"
#include "core/fmt/fmt_fprintf.h"
#include "std/os/file.h"
#include "std/string/string.h"
#include "std/string/string_writer.h"
#include "core/config_parser/parse_to_fields.h"

typedef struct mystruct
{
    string hello;
    string foo;
    string abc;
} mystruct;

mystruct mystruct_new()
{
    return (mystruct){
        .hello = string_new(),
        .foo = string_new(),
        .abc = string_new(),
    };
}

void mystruct_print(mystruct *ms)
{
    char h[256] = {0}, f[256] = {0}, a[256] = {0};
    string_copy_to_c(h, &ms->hello, sizeof(h));
    string_copy_to_c(f, &ms->foo, sizeof(f));
    string_copy_to_c(a, &ms->abc, sizeof(a));
    printf(
        "{'hello': '%s', 'foo': '%s', 'abc': '%s'}\n",
        h,
        f,
        a);
}

extern bool test_config_parser();

int main()
{
    if (!test_config_parser())
    {
        return 1;
    }
    mystruct ms = mystruct_new();
    fields f = FIELDS(
        FIELD("hello", string_writer(&ms.hello)),
        FIELD("foo", string_writer(&ms.foo)),
        FIELD("ABC", string_writer(&ms.abc)));
    reader src = reader_new(
        (void *)&STR_READER(STR_LIT("hello:world\nfoo:bar")),
        (read_func)str_reader_io_read);
    config_parser parser = config_parser_new(src, STR_ARR((char[2]){0}));

    config_parser_parse_to_fields_result res = config_parser_parse_to_fields(&parser, &f);
    if (res.status != config_parser_parse_to_fields_status_ok)
    {
        panic(
            "unexpected state: %s\n",
            config_parser_parse_to_fields_status_to_str(res.status).data);
    }

    mystruct_print(&ms);
    return 0;
}

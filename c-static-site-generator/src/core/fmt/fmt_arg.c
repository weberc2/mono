#include "core/fmt/fmt_arg.h"

size_t fmt_arg_format(fmt_arg arg, str buf)
{
    return arg.format(arg.data, buf);
}

size_t fmt_arg_str_format(fmt_arg_str *fas, str buf)
{
    size_t nc = str_copy_at(buf, fas->buffer, fas->cursor);
    if (nc < 1)
    {
        return 0;
    }
    fas->cursor += nc;
    return nc;
}

fmt_arg fmt_args_first(fmt_args *args)
{
    if (args->len > 0)
    {
        return args->data[0];
    }
    return (fmt_arg){NULL, NULL};
}

fmt_arg fmt_args_pop(fmt_args *args)
{
    if (args->len > 0)
    {
        fmt_arg ret = args->data[0];
        args->len--;
        args->data = &args->data[1];
        return ret;
    }

    return (fmt_arg){NULL, NULL};
}
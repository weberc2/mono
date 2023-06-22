#include "core/str/str.h"
#include "core/math/math.h"
#include "core/panic/panic.h"
#include <string.h>

str str_new(char *data, size_t len)
{
    str s;
    str_init(&s, data, len);
    return s;
}

void str_init(str *s, char *data, size_t len)
{
    s->data = data;
    s->len = len;
}

str str_slice(str s, size_t start, size_t end)
{
    if (start > s.len)
    {
        panic(
            "slicing str with len `%zu`: start index `%zu` is out of "
            "bounds",
            s.len,
            start);
    }
    else if (end > s.len)
    {
        panic(
            "slicing str with len `%zu`: end index `%zu` is out of "
            "bounds",
            s.len,
            end);
    }
    else if (start > end)
    {
        panic(
            "slicing str with len `%zu`: start index `%zu` exceeds end index "
            "`%zu`",
            s.len,
            start,
            end);
    }

    return str_new(s.data + start, end - start);
}

size_t str_copy(str dst, str src)
{
    size_t sz = min(dst.len, src.len);
    memmove(dst.data, src.data, sz);
    return sz;
}

size_t str_copy_at(str dst, str src, size_t start)
{
    return str_copy(dst, str_slice(src, start, src.len));
}

bool str_eq(str lhs, str rhs)
{
    return lhs.len == rhs.len && memcmp(lhs.data, rhs.data, lhs.len) == 0;
}

size_t str_copy_to_c(char *dst, str src, size_t len)
{
    size_t copied = min(len, src.len);
    memmove(dst, src.data, copied);
    dst[copied] = '\0';
    return copied;
}
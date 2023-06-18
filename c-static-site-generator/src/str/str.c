#include "str/str.h"
#include "math/math.h"
#include "panic/panic.h"
#include <string.h>

void str_init(str *s, char *data, size_t len)
{
    s->data = data;
    s->len = len;
}

void str_slice(str s, str *out, size_t start, size_t end)
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

    out->data = s.data + start;
    out->len = end - start;
}

size_t str_copy(str dst, str src)
{
    size_t sz = min(dst.len, src.len);
    memmove(dst.data, src.data, sz);
    return sz;
}

size_t str_copy_at(str dst, str src, size_t start)
{
    str tail;
    str_slice(src, &tail, start, src.len);
    return str_copy(dst, tail);
}

bool str_eq(str lhs, str rhs)
{
    return lhs.len == rhs.len && memcmp(lhs.data, rhs.data, lhs.len) == 0;
}
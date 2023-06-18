#include "byteslice/byteslice.h"
#include "math/math.h"
#include "panic/panic.h"
#include <string.h>

void byteslice_init(byteslice *bs, char *data, size_t len)
{
    bs->data = data;
    bs->len = len;
}

void byteslice_slice(byteslice bs, byteslice *out, size_t start, size_t end)
{
    if (start > bs.len)
    {
        panic(
            "slicing byteslice with len `%zu`: start index `%zu` is out of "
            "bounds",
            bs.len,
            start);
    }
    else if (end > bs.len)
    {
        panic(
            "slicing byteslice with len `%zu`: end index `%zu` is out of "
            "bounds",
            bs.len,
            end);
    }

    out->data = bs.data + start;
    out->len = end - start;
}

size_t byteslice_copy(byteslice dst, byteslice src)
{
    size_t sz = min(dst.len, src.len);
    memmove(dst.data, src.data, sz);
    return sz;
}

size_t byteslice_copy_at(byteslice dst, byteslice src, size_t start)
{
    byteslice tail;
    byteslice_slice(src, &tail, start, src.len);
    return byteslice_copy(dst, tail);
}

bool byteslice_eq(byteslice lhs, byteslice rhs)
{
    return lhs.len == rhs.len && memcmp(lhs.data, rhs.data, lhs.len) == 0;
}
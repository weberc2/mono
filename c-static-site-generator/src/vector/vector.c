#include "vector/vector.h"

#include <stdbool.h>

void vector_init(vector *v, size_t elt_size)
{
    v->data = NULL;
    v->len = 0;
    v->cap = 0;
    v->elt_size = elt_size;
}

void vector_init_with_cap(vector *v, size_t cap, size_t elt_size)
{
    v->data = calloc(cap, elt_size);
    v->len = 0;
    v->cap = cap;
    v->elt_size = elt_size;
}

void vector_drop(vector *v)
{
    free(v->data);
    v->cap = 0;
    v->len = 0;
}

void vector_grow(vector *v)
{
    void *old = v->data;
    v->cap = 2 * (v->cap + 1);
    v->data = calloc(v->cap, v->elt_size);
    memcpy(v->data, old, v->len * v->elt_size);
    free(old);
}

void *vector_alloc(vector *v)
{
    if (v->cap - v->len < 1)
    {
        vector_grow(v);
    }
    return v->data + (v->len++ * v->elt_size);
}

void vector_push(vector *v, void *value)
{
    void *dst = vector_alloc(v);
    memcpy(dst, value, v->elt_size);
}

bool vector_pop(vector *v, void *out)
{
    if (v->len < 1)
    {
        return false;
    }

    void *src = v->data + (v->elt_size * (v->len - 1));
    memcpy(out, src, v->elt_size);
    v->len--;
    return true;
}

void *vector_get(vector *v, size_t i)
{
    if (i >= v->len)
    {
        fprintf(
            stderr,
            "can't access index `%zu` in vector of len `%zu`",
            i,
            v->len);
        fflush(stderr);
        abort();
    }

    return v->data + (i * v->elt_size);
}
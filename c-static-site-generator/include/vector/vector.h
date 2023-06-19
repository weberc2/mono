#ifndef VECTOR_H
#define VECTOR_H

#include <stdbool.h>
#include <stddef.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

typedef struct
{
    void *data;
    size_t len;
    size_t cap;
    size_t elt_size;
} vector;

void vector_init(vector *v, size_t elt_size);
void vector_init_with_cap(vector *v, size_t cap, size_t elt_size);
void vector_drop(vector *v);
void vector_push(vector *v, void *value);
bool vector_pop(vector *v, void *out);
void *vector_get(vector *v, size_t i);
void *vector_alloc(vector *v);

#endif // VECTOR_H
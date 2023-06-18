#ifndef COPY_H
#define COPY_H

#include "reader.h"
#include "writer.h"
#include "error/error.h"

size_t copy(writer dst, reader src, errors *errs);

#endif // COPY_H
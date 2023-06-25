#include "core/io/eof.h"

static void __attribute__((constructor)) init()
{
    error_const((error *)(&eof), "end of file");
}
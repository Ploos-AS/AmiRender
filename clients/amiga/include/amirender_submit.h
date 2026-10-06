#ifndef AMIRENDER_SUBMIT_H
#define AMIRENDER_SUBMIT_H

#include <stddef.h>

struct amirender_job {
    const char *id;
    const char *engine;
    const char *scene;
    const char *output;
    int frame;
    int width;
    int height;
};

int amirender_build_submit(char *buffer, size_t size, const struct amirender_job *job);

#endif

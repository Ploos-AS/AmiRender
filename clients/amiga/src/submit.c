#include "amirender_submit.h"

#include <stdio.h>

static int safe_text(const char *s)
{
    const unsigned char *p = (const unsigned char *)s;

    if (s == NULL) {
        return 0;
    }
    while (*p != '\0') {
        if (*p < 0x20 || *p == '"' || *p == '\\') {
            return 0;
        }
        ++p;
    }
    return 1;
}

int amirender_build_submit(char *buffer, size_t size, const struct amirender_job *job)
{
    int n;

    if (buffer == NULL || size == 0 || job == NULL ||
        !safe_text(job->id) || !safe_text(job->engine) ||
        !safe_text(job->scene) || !safe_text(job->output)) {
        return -1;
    }

    n = snprintf(
        buffer, size,
        "{\"type\":\"SUBMIT\",\"job\":{\"id\":\"%s\",\"engine\":\"%s\","
        "\"scene\":\"%s\",\"output\":\"%s\",\"frame\":%d,\"width\":%d,\"height\":%d}}\n",
        job->id, job->engine, job->scene, job->output,
        job->frame, job->width, job->height);

    if (n < 0 || (size_t)n >= size) {
        return -1;
    }
    return n;
}

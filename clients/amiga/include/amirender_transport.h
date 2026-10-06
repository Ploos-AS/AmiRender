#ifndef AMIRENDER_TRANSPORT_H
#define AMIRENDER_TRANSPORT_H

#include <stddef.h>

struct amirender_job;

struct amirender_transport {
    void *context;
    int (*send)(void *context, const char *data, size_t length);
    int (*receive)(void *context, char *buffer, size_t size);
};

int amirender_submit_job(
    struct amirender_transport *transport,
    const struct amirender_job *job,
    char *reply,
    size_t reply_size);

#endif

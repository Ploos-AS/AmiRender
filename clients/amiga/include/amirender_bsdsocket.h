#ifndef AMIRENDER_BSDSOCKET_H
#define AMIRENDER_BSDSOCKET_H

#include "amirender_transport.h"

struct amirender_bsdsocket {
    int socket_fd;
};

int amirender_bsdsocket_connect(
    struct amirender_bsdsocket *socket,
    const char *host,
    unsigned short port,
    struct amirender_transport *transport);

void amirender_bsdsocket_close(struct amirender_bsdsocket *socket);

#endif

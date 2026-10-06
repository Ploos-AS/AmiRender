#include "amirender_submit.h"
#include "amirender_transport.h"

#include <string.h>

int amirender_submit_job(
    struct amirender_transport *transport,
    const struct amirender_job *job,
    char *reply,
    size_t reply_size)
{
    char request[1024];
    int length;
    int received;

    if (transport == NULL || transport->send == NULL || transport->receive == NULL ||
        reply == NULL || reply_size < 2) {
        return -1;
    }

    length = amirender_build_submit(request, sizeof(request), job);
    if (length < 0) {
        return -1;
    }
    if (transport->send(transport->context, request, (size_t)length) != length) {
        return -1;
    }

    received = transport->receive(transport->context, reply, reply_size - 1);
    if (received <= 0 || (size_t)received >= reply_size) {
        return -1;
    }
    reply[received] = '\0';

    if (strstr(reply, "\"type\":\"COMPLETE\"") == NULL) {
        return -1;
    }
    return 0;
}

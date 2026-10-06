/*
 * AmigaOS/m68k transport for AmiRender.
 *
 * This translation unit is intentionally not part of the host build. It is
 * compiled by the Amiga/Bebbo toolchain where bsdsocket.library headers and
 * stubs are available.
 */

#include "amirender_bsdsocket.h"

#include <sys/types.h>
#include <proto/bsdsocket.h>
#include <proto/exec.h>

#include <netdb.h>
#include <netinet/in.h>
#include <string.h>
#include <sys/socket.h>

struct Library *SocketBase = NULL;

static int socket_send(void *context, const char *data, size_t length)
{
    struct amirender_bsdsocket *state = (struct amirender_bsdsocket *)context;
    size_t sent = 0;

    while (sent < length) {
        int n = send(state->socket_fd, (void *)(data + sent), length - sent, 0);
        if (n <= 0) {
            return -1;
        }
        sent += (size_t)n;
    }
    return (int)sent;
}

static int socket_receive(void *context, char *buffer, size_t size)
{
    struct amirender_bsdsocket *state = (struct amirender_bsdsocket *)context;
    return recv(state->socket_fd, buffer, size, 0);
}

int amirender_bsdsocket_connect(
    struct amirender_bsdsocket *state,
    const char *host,
    unsigned short port,
    struct amirender_transport *transport)
{
    struct hostent *entry;
    struct sockaddr_in address;
    int fd;

    if (state == NULL || host == NULL || transport == NULL) {
        return -1;
    }

    SocketBase = OpenLibrary("bsdsocket.library", 4);
    if (SocketBase == NULL) {
        return -1;
    }

    entry = gethostbyname((char *)host);
    if (entry == NULL || entry->h_addr_list == NULL || entry->h_addr_list[0] == NULL) {
        CloseLibrary(SocketBase);
        SocketBase = NULL;
        return -1;
    }

    fd = socket(AF_INET, SOCK_STREAM, 0);
    if (fd < 0) {
        CloseLibrary(SocketBase);
        SocketBase = NULL;
        return -1;
    }

    memset(&address, 0, sizeof(address));
    address.sin_family = AF_INET;
    address.sin_port = htons(port);
    memcpy(&address.sin_addr, entry->h_addr_list[0], sizeof(address.sin_addr));

    if (connect(fd, (struct sockaddr *)&address, sizeof(address)) < 0) {
        CloseSocket(fd);
        CloseLibrary(SocketBase);
        SocketBase = NULL;
        return -1;
    }

    state->socket_fd = fd;
    transport->context = state;
    transport->send = socket_send;
    transport->receive = socket_receive;
    return 0;
}

void amirender_bsdsocket_close(struct amirender_bsdsocket *state)
{
    if (state != NULL && state->socket_fd >= 0) {
        CloseSocket(state->socket_fd);
        state->socket_fd = -1;
    }
    if (SocketBase != NULL) {
        CloseLibrary(SocketBase);
        SocketBase = NULL;
    }
}

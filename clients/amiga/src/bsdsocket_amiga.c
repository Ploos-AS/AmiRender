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

#include <netinet/in.h>
#include <string.h>
#include <sys/socket.h>

struct Library *SocketBase = NULL;


static int parse_ipv4(const char *text, unsigned long *result)
{
    unsigned long value = 0;
    int part;

    if (text == NULL || result == NULL) {
        return -1;
    }

    for (part = 0; part < 4; ++part) {
        unsigned long octet = 0;
        int digits = 0;

        while (*text >= '0' && *text <= '9') {
            octet = octet * 10UL + (unsigned long)(*text - '0');
            if (octet > 255UL) {
                return -1;
            }
            ++text;
            ++digits;
        }
        if (digits == 0) {
            return -1;
        }
        value = (value << 8) | octet;
        if (part < 3) {
            if (*text != '.') {
                return -1;
            }
            ++text;
        }
    }
    if (*text != '\0') {
        return -1;
    }

    *result = value;
    return 0;
}

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
    size_t received = 0;

    if (buffer == NULL || size == 0) {
        return -1;
    }

    while (received < size) {
        int n = recv(state->socket_fd, buffer + received, size - received, 0);
        size_t i;

        if (n <= 0) {
            return received > 0 ? (int)received : -1;
        }

        for (i = 0; i < (size_t)n; ++i) {
            if (buffer[received + i] == '\n') {
                return (int)(received + i + 1);
            }
        }
        received += (size_t)n;
    }

    return (int)received;
}

int amirender_bsdsocket_connect(
    struct amirender_bsdsocket *state,
    const char *host,
    unsigned short port,
    struct amirender_transport *transport)
{
    struct sockaddr_in address;
    int fd;
    unsigned long ipv4;

    if (state == NULL || host == NULL || transport == NULL) {
        return -1;
    }

    SocketBase = OpenLibrary("bsdsocket.library", 4);
    if (SocketBase == NULL) {
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
    if (parse_ipv4(host, &ipv4) != 0) {
        CloseSocket(fd);
        CloseLibrary(SocketBase);
        SocketBase = NULL;
        return -1;
    }
    address.sin_addr.s_addr = htonl(ipv4);

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

#include "amirender_submit.h"
#include "amirender_transport.h"

#include <stdio.h>
#include <string.h>

struct fake_transport {
    char sent[1024];
};

static int fake_send(void *context, const char *data, size_t length)
{
    struct fake_transport *fake = (struct fake_transport *)context;

    if (length >= sizeof(fake->sent)) {
        return -1;
    }
    memcpy(fake->sent, data, length);
    fake->sent[length] = '\0';
    return (int)length;
}

static int fake_receive(void *context, char *buffer, size_t size)
{
    const char *reply =
        "{\"type\":\"COMPLETE\",\"job_id\":\"amiga-0002\","
        "\"engine\":\"povray\",\"output\":\"RAM:frame.png\"}\n";
    size_t length = strlen(reply);

    (void)context;
    if (length > size) {
        return -1;
    }
    memcpy(buffer, reply, length);
    return (int)length;
}

int main(void)
{
    struct fake_transport fake = {{0}};
    struct amirender_transport transport = {&fake, fake_send, fake_receive};
    struct amirender_job job = {
        "amiga-0002", "povray", "DH0:Scenes/demo.pov",
        "RAM:frame.png", 0, 320, 256
    };
    char reply[512];

    if (amirender_submit_job(&transport, &job, reply, sizeof(reply)) != 0) {
        return 1;
    }
    if (strstr(fake.sent, "\"id\":\"amiga-0002\"") == NULL) {
        fprintf(stderr, "job id missing from request\n");
        return 2;
    }
    if (strstr(reply, "\"type\":\"COMPLETE\"") == NULL) {
        fprintf(stderr, "completion missing from reply\n");
        return 3;
    }
    return 0;
}

#include "amirender_submit.h"
#include "amirender_transport.h"

#include <stdio.h>
#include <string.h>

struct fake_transport {
    char sent[1024];
    const char *reply;
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
    struct fake_transport *fake = (struct fake_transport *)context;
    const char *reply = fake->reply;
    size_t length = strlen(reply);

    if (length > size) {
        return -1;
    }
    memcpy(buffer, reply, length);
    return (int)length;
}

int main(void)
{
    struct fake_transport fake = {{0}, NULL};
    struct amirender_transport transport = {&fake, fake_send, fake_receive};
    struct amirender_job job = {
        "amiga-0002", "povray", "DH0:Scenes/demo.pov",
        "RAM:frame.png", 0, 320, 256
    };
    char reply[512];
    char asset[256];
    const unsigned char scene[] = "camera {}\n";

    fake.reply = "{\"type\":\"COMPLETE\",\"job_id\":\"amiga-0002\","
                 "\"engine\":\"povray\",\"output\":\"RAM:frame.png\"}\n";
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

    fake.reply = "{\"type\":\"STAGED\",\"asset\":\"/tmp/amirender/scene.pov\"}\n";
    if (amirender_upload_asset(
            &transport, "scene.pov", scene, sizeof(scene) - 1, asset, sizeof(asset)) != 0) {
        return 4;
    }
    if (strcmp(asset, "/tmp/amirender/scene.pov") != 0) {
        fprintf(stderr, "unexpected staged asset: %s\n", asset);
        return 5;
    }
    if (strstr(fake.sent, "\"type\":\"UPLOAD\"") == NULL ||
        strstr(fake.sent, "Y2FtZXJhIHt9Cg==") == NULL) {
        fprintf(stderr, "upload request missing expected payload\n");
        return 6;
    }

    fake.reply = "{\"type\":\"FAILED\",\"error\":\"staging failed\"}\n";
    if (amirender_upload_asset(
            &transport, "scene.pov", scene, sizeof(scene) - 1, asset, sizeof(asset)) == 0) {
        return 7;
    }
    return 0;
}

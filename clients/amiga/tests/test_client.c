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
    unsigned char downloaded[16];
    size_t downloaded_size = 0;
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

    fake.reply = "{\"type\":\"DATA\",\"asset\":\"/tmp/amirender-asset-x/frame.png\","
                 "\"data\":\"iVBORw0K\"}\n";
    if (amirender_download_asset(
            &transport, "/tmp/amirender-asset-x/frame.png",
            downloaded, sizeof(downloaded), &downloaded_size) != 0) {
        return 8;
    }
    if (downloaded_size != 6 ||
        downloaded[0] != 0x89 || downloaded[1] != 'P' ||
        downloaded[2] != 'N' || downloaded[3] != 'G' ||
        downloaded[4] != 0x0d || downloaded[5] != 0x0a) {
        fprintf(stderr, "downloaded data mismatch\n");
        return 9;
    }
    if (strstr(fake.sent, "\"type\":\"DOWNLOAD\"") == NULL) {
        fprintf(stderr, "download request missing\n");
        return 10;
    }

    fake.reply = "{\"type\":\"FAILED\",\"error\":\"asset unavailable\"}\n";
    if (amirender_download_asset(
            &transport, "/tmp/amirender-asset-x/frame.png",
            downloaded, sizeof(downloaded), &downloaded_size) == 0) {
        return 11;
    }
    fake.reply = "{\"type\":\"COMPLETE\",\"output\":\"/tmp/old.png\",\"asset_id\":\"asset-new\"}";
    if (amirender_extract_output_asset(fake.reply, asset, sizeof(asset)) != 0 ||
        strcmp(asset, "asset-new") != 0) {
        fprintf(stderr, "opaque asset ID not preferred\n");
        return 12;
    }
    fake.reply = "{\"type\":\"COMPLETE\",\"output\":\"/tmp/old.png\"}";
    if (amirender_extract_output_asset(fake.reply, asset, sizeof(asset)) != 0 ||
        strcmp(asset, "/tmp/old.png") != 0) {
        fprintf(stderr, "legacy output fallback failed\n");
        return 13;
    }
    fake.reply = "{\"type\":\"COMPLETE\",\"output\":\"/tmp/old.png\",\"asset_id\":\"\"}";
    if (amirender_extract_output_asset(fake.reply, asset, sizeof(asset)) != 0 ||
        strcmp(asset, "/tmp/old.png") != 0) {
        fprintf(stderr, "empty asset ID fallback failed\n");
        return 14;
    }
    fake.reply = "{\"type\":\"STAGING\",\"id\":\"asset-opaque\",\"asset\":\"/tmp/scene.pov\"}";
    if (amirender_upload_begin(&transport, "scene.pov", 10, asset, sizeof(asset)) != 0 ||
        strcmp(asset, "asset-opaque") != 0) {
        fprintf(stderr, "opaque upload ID not preferred\n");
        return 15;
    }
    fake.reply = "{\"type\":\"STAGING\",\"asset\":\"/tmp/scene.pov\"}";
    if (amirender_upload_begin(&transport, "scene.pov", 10, asset, sizeof(asset)) != 0 ||
        strcmp(asset, "/tmp/scene.pov") != 0) {
        fprintf(stderr, "legacy upload asset fallback failed\n");
        return 16;
    }
    fake.reply = "{\"type\":\"STAGING\",\"id\":\"asset-roundtrip\",\"asset\":\"/tmp/scene.pov\"}";
    if (amirender_upload_begin(&transport, "scene.pov", 10, asset, sizeof(asset)) != 0 ||
        strcmp(asset, "asset-roundtrip") != 0) return 17;
    fake.reply = "{\"type\":\"CHUNKED\",\"offset\":10}";
    if (amirender_upload_chunk(&transport, asset, 0, scene, sizeof(scene) - 1) != 0 ||
        strstr(fake.sent, "\"asset\":\"asset-roundtrip\"") == NULL) {
        fprintf(stderr, "upload chunk did not use opaque ID\n");
        return 18;
    }
    fake.reply = "{\"type\":\"STAGED\"}";
    if (amirender_upload_end(&transport, asset, 10) != 0 ||
        strstr(fake.sent, "\"asset\":\"asset-roundtrip\"") == NULL) {
        fprintf(stderr, "upload end did not use opaque ID\n");
        return 19;
    }
    {
        int eof = 0;
        size_t count = 0;
        fake.reply = "{\"type\":\"DATA\",\"data\":\"iVBORw0K\",\"eof\":false}";
        if (amirender_download_chunk(&transport, "asset-rendered", 0,
                downloaded, sizeof(downloaded), &count, &eof) != 0 ||
            count != 6 || eof != 0 ||
            strstr(fake.sent, "\"asset\":\"asset-rendered\"") == NULL ||
            strstr(fake.sent, "\"offset\":0") == NULL ||
            downloaded[0] != 0x89 || downloaded[1] != 'P') {
            fprintf(stderr, "first opaque download chunk failed\n");
            return 20;
        }
        fake.reply = "{\"type\":\"DATA\",\"data\":\"Tkc=\",\"eof\":true}";
        if (amirender_download_chunk(&transport, "asset-rendered", 6,
                downloaded, sizeof(downloaded), &count, &eof) != 0 ||
            count != 2 || eof != 1 ||
            strstr(fake.sent, "\"asset\":\"asset-rendered\"") == NULL ||
            strstr(fake.sent, "\"offset\":6") == NULL ||
            downloaded[0] != 'N' || downloaded[1] != 'G') {
            fprintf(stderr, "final opaque download chunk failed\n");
            return 21;
        }
        fake.reply = "{\"type\":\"DATA\",\"eof\":true}";
        if (amirender_download_chunk(&transport, "asset-rendered", 8,
                downloaded, sizeof(downloaded), &count, &eof) != 0 ||
            count != 0 || eof != 1) {
            fprintf(stderr, "empty EOF response failed\n");
            return 22;
        }
    }
    return 0;
}

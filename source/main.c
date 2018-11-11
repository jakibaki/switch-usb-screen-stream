#include <switch.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <malloc.h>

size_t transport_safe_read(void *buffer, size_t size)
{
    u8 *bufptr = buffer;
    size_t cursize = size;
    size_t tmpsize = 0;

    while (cursize)
    {
        tmpsize = usbCommsRead(bufptr, cursize);
        bufptr += tmpsize;
        cursize -= tmpsize;
    }

    return size;
}

size_t transport_safe_write(const void *buffer, size_t size)
{
    const u8 *bufptr = (const u8 *)buffer;
    size_t cursize = size;
    size_t tmpsize = 0;

    while (cursize)
    {
        tmpsize = usbCommsWrite(bufptr, cursize);
        bufptr += tmpsize;
        cursize -= tmpsize;
    }

    return size;
}

struct JoyPkg {
    unsigned long heldKeys;
    int lJoyX;
    int lJoyY;
    int rJoyX;
    int rJoyY;
};

void inputPoller(void* DISCARD) {
    JoystickPosition lJoy;
    JoystickPosition rJoy;
    struct JoyPkg* pkg = memalign(0x1000, sizeof(struct JoyPkg));
    while(appletMainLoop()) {
        pkg->heldKeys = hidKeysHeld(CONTROLLER_P1_AUTO);
        hidJoystickRead(&lJoy, CONTROLLER_P1_AUTO, JOYSTICK_LEFT);
        hidJoystickRead(&rJoy, CONTROLLER_P1_AUTO, JOYSTICK_RIGHT);
        pkg->lJoyX = lJoy.dx;
        pkg->lJoyY = lJoy.dy;
        pkg->rJoyX = rJoy.dx;
        pkg->rJoyY = rJoy.dy;
        transport_safe_write(pkg, sizeof(struct JoyPkg));
        svcSleepThread(33333333);
    }
}

int main(int argc, char **argv)
{
    u32* framebuf;
    u32  cnt=0;

    Thread inputThread;
    threadCreate(&inputThread, inputPoller, NULL, 0x1000, 0x3B, -2);
    threadStart(&inputThread);

    gfxInitDefault();
    gfxConfigureResolution(640, 360);


    u8* imageptr = memalign(0x1000, 640 * 360 * 4);
    

    usbCommsInitialize();


    while(appletMainLoop())
    {
        hidScanInput();
        
        u32 width, height;
        u32 pos;
        framebuf = (u32*) gfxGetFramebuffer((u32*)&width, (u32*)&height);


        transport_safe_read(imageptr, 640 * 360 * 4);

        //Each pixel is 4-bytes due to RGBA8888.
        u32 x, y;
        for (y=0; y<height; y++)//Access the buffer linearly.
        {
            for (x=0; x<width; x++)
            {
                pos = y * width + x;
                framebuf[pos] = RGBA8_MAXALPHA(imageptr[pos*4+0]+(cnt*4), imageptr[pos*4+1], imageptr[pos*4+2]);
            }
        }
        gfxFlushBuffers();
        gfxSwapBuffers();
    }
    usbCommsExit();

    gfxExit();
    return 0;
}

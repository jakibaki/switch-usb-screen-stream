#include <switch.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <malloc.h>

#ifdef DISPLAY_IMAGE
#include "image_bin.h"//Your own raw RGB888 1280x720 image at "data/image.bin" is required.
#endif

//See also libnx gfx.h.


#define MIN(a,b) (((a)<(b))?(a):(b))
#define MAX(a,b) (((a)>(b))?(a):(b))
size_t transport_safe_read(void *buffer, size_t size)
{
    u8 *bufptr = buffer;
    size_t cursize = size;
    size_t tmpsize = 0;

    while (cursize)
    {
        printf("%ld\n", cursize);
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



int main(int argc, char **argv)
{
    u32* framebuf;
    u32  cnt=0;

    //Enable max-1080p support. Remove for 720p-only resolution.
    //gfxInitResolutionDefault();

    gfxInitDefault();
    //gfxInitResolution(640, 360);
    gfxConfigureResolution(640, 360);


    u8* imageptr = memalign(0x1000, 640 * 360 * 4);
    

    //Set current resolution automatically depending on current/changed OperationMode. Only use this when using gfxInitResolution*().
    //gfxConfigureAutoResolutionDefault(true);

    Result res = usbCommsInitialize();


    while(appletMainLoop())
    {
        //Scan all the inputs. This should be done once for each frame
        hidScanInput();

        //hidKeysDown returns information about which buttons have been just pressed (and they weren't in the previous frame)
        u64 kDown = hidKeysDown(CONTROLLER_P1_AUTO);

        if (kDown & KEY_PLUS) break; // break in order to return to hbmenu


        
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

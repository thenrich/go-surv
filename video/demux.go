package video

import (
	"bytes"
	"fmt"
	"github.com/giorgisio/goav/avcodec"
	"github.com/giorgisio/goav/avformat"
	"github.com/giorgisio/goav/avutil"
	"github.com/giorgisio/goav/swscale"
	"github.com/pkg/errors"
	"image"
	"image/jpeg"
	"log"
	"os"
	"unsafe"
)



type demuxer struct {
	ctx *avformat.Context
	codecCtxOrig *avformat.CodecContext
	codecCtx *avformat.CodecContext

	// video stream index
	videoStreamIndex int
}

func (d *demuxer) Read() error {
	return d.readFrame()
}

func (d *demuxer) readFrames() ([]image.Image, error) {
	// Read frames and save first five frames to disk
	frameNumber := 1
	packet := avcodec.AvPacketAlloc()
	for d.ctx.AvReadFrame(packet) >= 0 {
		// Is this a packet from the video stream?
		if packet.StreamIndex() == d.videoStreamIndex {
			// Decode video frame
			response := d.codecCtx.AvcodecSendPacket(packet)
			fmt.Println(response)
			fmt.Println("next...")
			if response < 0 {
				return errors.Errorf("error sending packet to decoder: %s\n", avutil.ErrorFromCode(response))
			}
			for response >= 0 {

				// Allocate video frame
				pFrame := avutil.AvFrameAlloc()

				// Allocate an AVFrame structure
				pFrameRGB := avutil.AvFrameAlloc()
				if pFrameRGB == nil {
					return errors.Errorf("unable to create frame\n")
				}

				// Determine required buffer size and allocate buffer
				numBytes := uintptr(avcodec.AvpictureGetSize(avcodec.AV_PIX_FMT_RGB24, d.codecCtx.Width(),
					d.codecCtx.Height()))
				buffer := avutil.AvMalloc(numBytes)

				// Assign appropriate parts of buffer to image planes in pFrameRGB
				// Note that pFrameRGB is an AVFrame, but AVFrame is a superset
				// of AVPicture
				avp := (*avcodec.Picture)(unsafe.Pointer(pFrameRGB))
				avp.AvpictureFill((*uint8)(buffer), avcodec.AV_PIX_FMT_RGB24, d.codecCtx.Width(), d.codecCtx.Height())

				// initialize SWS context for software scaling
				swsCtx := swscale.SwsGetcontext(
					d.codecCtx.Width(),
					d.codecCtx.Height(),
					(swscale.PixelFormat)(d.codecCtx.PixFmt()),
					d.codecCtx.Width(),
					d.codecCtx.Height(),
					avcodec.AV_PIX_FMT_RGB24,
					avcodec.SWS_BILINEAR,
					nil,
					nil,
					nil,
				)

				response = d.codecCtx.AvcodecReceiveFrame((*avcodec.Frame)(unsafe.Pointer(pFrame)))
				fmt.Println(response)
				if response == avutil.AvErrorEAGAIN || response == avutil.AvErrorEOF || response == -11 {
					break
				} else if response < 0 {
					return errors.Errorf("error while receiving a frame from the decoder: %s\n", avutil.ErrorFromCode(response))
				}

				// Convert the image from its native format to RGB
				swscale.SwsScale2(swsCtx, avutil.Data(pFrame),
					avutil.Linesize(pFrame), 0, d.codecCtx.Height(),
					avutil.Data(pFrameRGB), avutil.Linesize(pFrameRGB))

				// Save the frame to disk
				fmt.Printf("Writing frame %d\n", frameNumber)
				img, err := avutil.GetPictureRGB(pFrameRGB)
				if err != nil {
					return errors.Wrap(err, "error getting picture from frame")
				}

			}
		}

		// Free the packet that was allocated by av_read_frame
		packet.AvFreePacket()
	}
}

func (d *demuxer) close() error {
	// Close the codecs
	d.codecCtx.AvcodecClose()
	(*avcodec.Context)(unsafe.Pointer(&d.codecCtxOrig)).AvcodecClose()

	// Close the video file
	d.ctx.AvformatCloseInput()

	return nil
}

func (d *demuxer) open() error {
	// Find the first video stream
	for i := 0; i < int(d.ctx.NbStreams()); i++ {
		switch d.ctx.Streams()[i].CodecParameters().AvCodecGetType() {
		case avformat.AVMEDIA_TYPE_VIDEO:
			// Get a pointer to the codec context for the video stream
			d.codecCtxOrig = d.ctx.Streams()[i].Codec()
			// Find the decoder for the video stream
			pCodec := avcodec.AvcodecFindDecoder(avcodec.CodecId(d.codecCtxOrig.GetCodecId()))
			if pCodec == nil {
				return errors.Errorf("unsupported codec")
			}
			// Copy context
			d.codecCtx = pCodec.AvcodecAllocContext3()
			if d.codecCtx.AvcodecCopyContext((*avcodec.Context)(unsafe.Pointer(&d.codecCtxOrig))) != 0 {
				return errors.Errorf("couldn't copy codec context")
			}
			// Open codec
			if d.codecCtx.AvcodecOpen2(pCodec, nil) < 0 {
				return errors.Errorf("couldn't open codec")
			}

			d.videoStreamIndex = i
			return nil

		}

	}

	return errors.Errorf("couldn't find video stream")
}
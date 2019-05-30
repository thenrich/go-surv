package video

import (
	"github.com/giorgisio/goav/avcodec"
	"github.com/giorgisio/goav/avformat"
	"github.com/pkg/errors"
	"unsafe"
)

type demuxer struct {
	ctx *avformat.Context
	codecCtxOrig *avformat.CodecContext
	codecCtx *avformat.CodecContext
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

			return nil

		}

	}

	return errors.Errorf("couldn't find video stream")
}
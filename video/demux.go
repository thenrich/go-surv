package video

import (
	"github.com/3d0c/gmf"
	"github.com/pkg/errors"
	"io"
	"log"
)

type demuxer struct {
	// URL to read from
	url string

	inputCtx    *gmf.FmtCtx
	srcVideo    *gmf.Stream
	inputStream *gmf.Stream
	codecCtx    *gmf.CodecCtx

	// video stream index
	videoStreamIndex int

	// Images
	imgCodecCtx *gmf.CodecCtx
	imgSwsCtx   *gmf.SwsCtx

	// Video
}

func (d *demuxer) ReadFrames() ([]*gmf.Frame, error) {
	// Loop until we get a good pkt
	//var err error
	var pkt *gmf.Packet
	var frames []*gmf.Frame
	for {
		var err error
		pkt, err = d.inputCtx.GetNextPacket()
		if err != nil && err != io.EOF {
			if pkt != nil {
				pkt.Free()
			}

			log.Println(errors.Wrap(err, "error getting packet, continue"))
			continue
		} else if err != nil && pkt == nil {
			log.Println(errors.Wrap(err, "error and nil packet, continue"))
			continue
		}

		if err == io.EOF {
			log.Println(errors.Wrap(err, "reached EOF"))
			return nil, err
		}

		if pkt != nil && pkt.StreamIndex() != d.srcVideo.Index() {
			log.Println("pkt from wrong stream, continue")
			continue
		}

		if pkt == nil {
			log.Println("nil packet after read frames, continue")
			continue
		}

		frames, err = d.inputStream.CodecCtx().Decode(pkt)
		if err != nil {
			log.Println(errors.Wrap(err, "fatal error during decoding, continue"))
			continue

		}
		if len(frames) == 0 {
			continue
		}

		break
	}

	//if frames, err = gmf.DefaultRescaler(d.imgSwsCtx, frames); err != nil {
	//	return nil, errors.Wrap(err, "error rescaling")
	//}


	if pkt != nil {
		pkt.Free()
		pkt = nil
	}

	return frames, nil

}

func (d *demuxer) open() error {
	ctx, err := gmf.NewInputCtx(d.url)
	if err != nil {
		return errors.Wrapf(err, "error opening %s\n", d.url)
	}

	srcVideo, err := ctx.GetBestStream(gmf.AVMEDIA_TYPE_VIDEO)
	if err != nil {
		return errors.Wrapf(err, "error finding stream\n")
	}

	inputStream, err := ctx.GetStream(srcVideo.Index())
	if err != nil {
		return errors.Wrap(err, "error getting stream")
	}

	d.inputCtx = ctx
	d.srcVideo = srcVideo
	d.videoStreamIndex = srcVideo.Index()
	d.inputStream = inputStream
	d.codecCtx = srcVideo.CodecCtx()



	return nil
}

func (d *demuxer) Close() error {
	d.inputCtx.Free()
	d.inputStream.Free()
	d.imgCodecCtx.Free()
	d.imgSwsCtx.Free()
	gmf.Release(d.imgCodecCtx)

	return nil
}

func NewDemuxer(url string) *demuxer {
	return &demuxer{
		url: url,
	}
}

func setEncoderParams(cc *gmf.CodecCtx, width int, height int, experimental bool) {
	cc.SetTimeBase(gmf.AVR{Num: 1, Den: 1000})
	cc.SetPixFmt(gmf.AV_PIX_FMT_RGB24).SetWidth(width).SetHeight(height)
	if experimental {
		cc.SetStrictCompliance(gmf.FF_COMPLIANCE_EXPERIMENTAL)
	}
}

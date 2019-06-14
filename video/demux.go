package video

import (
	"github.com/3d0c/gmf"
	"github.com/pkg/errors"
	"io"
	"log"
)



type demuxer struct {
	url string

	inputCtx *gmf.FmtCtx
	srcVideo *gmf.Stream
	inputStream *gmf.Stream
	codecCtx *gmf.CodecCtx


	// video stream index
	videoStreamIndex int

	// Images
	imgCodecCtx *gmf.CodecCtx
	imgSwsCtx *gmf.SwsCtx

	// Video
}

func (d *demuxer) Read() error {
	drain := -1
	frameCount := 0
	for {
		if drain >= 0 {
			break
		}

		pkt, err := d.inputCtx.GetNextPacket()
		if err != nil && err != io.EOF {
			if pkt != nil {
				pkt.Free()
			}
			log.Printf("error getting next packet - %s", err)
			break
		} else if err != nil && pkt == nil {
			drain = 0
		}

		if pkt != nil && pkt.StreamIndex() != d.srcVideo.Index() {
			continue
		}

		frames, err := d.inputStream.CodecCtx().Decode(pkt)
		if err != nil {
			log.Printf("Fatal error during decoding - %s", err)
			break
		}

		// Decode() method doesn't treat EAGAIN and EOF as errors
		// it returns empty frames slice instead. Countinue until
		// input EOF or frames received.
		if len(frames) == 0 && drain < 0 {
			continue
		}

		if frames, err = gmf.DefaultRescaler(d.imgSwsCtx, frames); err != nil {
			log.Println(errors.Wrap(err, "error rescaling").Error())
		}

		// use encoder here
		log.Println("Encode something")
		d.imgCodecCtx.Encode()


		for i := range frames {
			frames[i].Free()
			frameCount++
		}

		if pkt != nil {
			pkt.Free()
			pkt = nil
		}

	}
}

func (d *demuxer) open() error {
	ctx, err := gmf.NewInputCtx(d.url)
	if err != nil {
		return errors.Wrapf(err, "error opening %s\n", s.cam.SourceURL)
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

	// Add image encoder
	codec, err := gmf.FindEncoder("jpg")
	if err != nil {
		return errors.Wrap(err, "error finding jpeg encoder")
	}

	imgCodecCtx := gmf.NewCodecCtx(codec)
	setEncoderParams(imgCodecCtx, srcVideo.CodecCtx().Width(), srcVideo.CodecCtx().Height(), codec.IsExperimental())
	if err := imgCodecCtx.Open(nil); err != nil {
		return errors.Wrap(err, "error opening img codec")
	}

	d.imgCodecCtx = imgCodecCtx
	imgCodecCtx.Encode()

	var imgSwsCtx *gmf.SwsCtx
	if imgSwsCtx, err = gmf.NewSwsCtx(
		d.codecCtx.Width(),
		d.codecCtx.Height(),
		d.codecCtx.PixFmt(),
		imgCodecCtx.Width(),
		imgCodecCtx.Height(),
		imgCodecCtx.PixFmt(),
		gmf.SWS_BICUBIC); err != nil {
		return errors.Wrap(err, "error creating sws context")
	}

	d.imgSwsCtx = imgSwsCtx


	return nil
}

func (d *demuxer) close() {
	d.inputCtx.Free()
	d.inputStream.Free()
	d.imgCodecCtx.Free()
	d.imgSwsCtx.Free()
	gmf.Release(d.imgCodecCtx)
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
package video

import (
	"bytes"
	"github.com/3d0c/gmf"
	"github.com/pkg/errors"
	"image"
	"image/jpeg"
	"time"
)

// Writer defines the interface for writing video packets
type Writer interface {
	//Open(streams []av.CodecData) error
	Write(writer []*gmf.Frame) error
	SetCodecContext(ctx *gmf.CodecCtx) error
	Close() error
}

//
//// CloudWriter defines the interface for writing files to cloud providers
//type CloudWriter interface {
//	UploadFile(r io.ReadCloser, key string) error
//}
//
//type CloudStorage struct {
//	*LocalWriter
//	writer CloudWriter
//
//	nextUploadFile string
//	nextUploadTime time.Time
//}
//
//func (cs *CloudStorage) Write(pkt av.Packet) error {
//	if err := cs.LocalWriter.dst.WritePacket(pkt); err != nil {
//		return errors.Wrap(err, "error writing packet to local writer")
//	}
//
//	if cs.nextRotation.Before(time.Now().UTC()) {
//		log.Println("Time to rotate")
//		// Set file to upload
//		cs.nextUploadFile = cs.outfile
//		cs.nextUploadTime = cs.now
//		if err := cs.rotate(); err != nil {
//			log.Println(errors.Wrap(err, "error rotating"))
//		}
//	}
//
//	return nil
//}
//
//// Rotate closes the local writer and reopens it at the current time
//func (cs *CloudStorage) rotate() error {
//	if err := cs.Close(); err != nil {
//		return errors.Wrap(err, "rotate: error closing local writer")
//	}
//
//	if err := cs.Open(nil); err != nil {
//		return errors.Wrap(err, "rotate: error opening new local writer")
//	}
//
//	cs.nextRotation = time.Now().UTC().Add(cs.duration)
//
//	return nil
//}
//
//// Close begins uploading completed file to S3
//func (cs *CloudStorage) Close() error {
//	if err := cs.LocalWriter.Close(); err != nil {
//		log.Println(errors.Wrap(err, "error closing file in s3 writer"))
//	}
//
//	if cs.nextUploadFile != "" {
//		go func() {
//			f, err := os.Open(cs.nextUploadFile)
//			if err != nil {
//				log.Println(err)
//				return
//			}
//			defer f.Close()
//			key := fmt.Sprintf("%d-%d-%d/%s", cs.nextUploadTime.Year(), cs.nextUploadTime.Month(), cs.nextUploadTime.Day(), strings.Replace(cs.nextUploadFile, "/tmp", "", 1))
//			err = cs.writer.UploadFile(f, key)
//			if err != nil {
//				log.Println(err)
//			}
//			// Copy our reference to the filename so we can clear nextUploadFile
//			deleteFile := cs.nextUploadFile
//			cs.nextUploadFile = ""
//			if err := os.Remove(deleteFile); err != nil {
//				log.Println(err)
//				return
//			}
//
//		}()
//	}
//
//	return nil
//}
//
//// NewCloudStorage creates a new CloudStorage instance with the given configuration and CloudWriter
//func NewCloudStorage(name string, interval time.Duration, cfg *config.Config, cloud CloudWriter) *CloudStorage {
//	localPath := fmt.Sprintf("/tmp/%s", name)
//	return &CloudStorage{LocalWriter: NewLocalWriter(localPath, interval), writer: cloud}
//}
//
//type LocalWriter struct {
//	name         string
//	outfile      string
//	dst          av.MuxCloser
//	duration     time.Duration
//	nextRotation time.Time
//	now          time.Time
//	streams      []av.CodecData
//}
//
//func (lw *LocalWriter) Open(streams []av.CodecData) error {
//	// Create timestamp for the current time
//	lw.now = time.Now().UTC()
//	lw.outfile = lw.filename(lw.name, lw.now)
//
//	// Set streams
//	if streams != nil {
//		lw.streams = streams
//	}
//
//	dst, err := avutil.Create(lw.outfile)
//	if err != nil {
//		return errors.Wrap(err, "error creating dest")
//	}
//	lw.dst = dst
//
//	if err := lw.dst.WriteHeader(lw.streams); err != nil {
//		return errors.Wrap(err, "error writing header for local writer")
//	}
//
//	return nil
//}
//
//// Rotate closes the local writer and reopens it at the current time
//func (lw *LocalWriter) rotate() error {
//	if err := lw.Close(); err != nil {
//		return errors.Wrap(err, "rotate: error closing local writer")
//	}
//
//	if err := lw.Open(nil); err != nil {
//		return errors.Wrap(err, "rotate: error opening new local writer")
//	}
//
//	lw.nextRotation = time.Now().UTC().Add(lw.duration)
//
//	return nil
//}
//
//func (lw *LocalWriter) filename(name string, t time.Time) string {
//	return fmt.Sprintf("%s-%d-%d-%d-%d-%d.mp4", name, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
//}
//
//func (lw *LocalWriter) Close() error {
//	if err := lw.dst.WriteTrailer(); err != nil {
//		return errors.Wrap(err, "error writing trailer to local writer")
//	}
//	if err := lw.dst.Close(); err != nil {
//		return errors.Wrap(err, "error closing local writer")
//	}
//
//	return nil
//}
//
//func (lw *LocalWriter) Write(pkt av.Packet) error {
//	if err := lw.dst.WritePacket(pkt); err != nil {
//		return errors.Wrap(err, "error writing packet to local writer")
//	}
//
//	if lw.nextRotation.Before(time.Now().UTC()) {
//		log.Println("Time to rotate")
//		if err := lw.rotate(); err != nil {
//			log.Println(errors.Wrap(err, "error rotating"))
//		}
//	}
//
//	return nil
//}
//
//// NewLocalWriter creates a new writer for storing videos locally
//func NewLocalWriter(name string, interval time.Duration) *LocalWriter {
//	return &LocalWriter{name: name, duration: interval, nextRotation: time.Now().UTC().Add(interval)}
//
//}
//
type StillWriter struct {
	// FFmpeg decoder
	//videoDecoder *ffmpeg.VideoDecoder
	codecCtx *gmf.CodecCtx
	timebase gmf.AVRational

	// Still channel
	stills chan *Still

	lastStill int64
}

//func (sw *StillWriter) Open(streams []av.CodecData) error {
//	return nil
//}

func (sw *StillWriter) SetCodecContext(ctx *gmf.CodecCtx) error {
	sw.codecCtx = ctx
	return nil
}

func (sw *StillWriter) SetTimeBase(tb gmf.AVRational) {
	sw.timebase = tb

}

func (sw *StillWriter) Close() error {
	return nil
}

func (sw *StillWriter) Write(frames []*gmf.Frame) error {
	codec, err := gmf.FindEncoder("png")
	if err != nil {
		return errors.Wrap(err, "error finding encoder")
	}

	cc := gmf.NewCodecCtx(codec)
	defer gmf.Release(cc)

	cc.SetTimeBase(sw.timebase.AVR())
	cc.SetPixFmt(
		gmf.AV_PIX_FMT_RGB24).SetWidth(
		sw.codecCtx.Width()).SetHeight(sw.codecCtx.Height())

	if codec.IsExperimental() {
		cc.SetStrictCompliance(gmf.FF_COMPLIANCE_EXPERIMENTAL)
	}

	if err := cc.Open(nil); err != nil {
		return errors.Wrap(err, "error opening codec")
	}

	var swsCtx *gmf.SwsCtx
	if swsCtx, err = gmf.NewSwsCtx(
		sw.codecCtx.Width(),
		sw.codecCtx.Height(),
		sw.codecCtx.PixFmt(),
		cc.Width(),
		cc.Height(),
		cc.PixFmt(),
		gmf.SWS_BICUBIC); err != nil {
		return errors.Wrap(err, "error create sws ctx")
	}

	defer swsCtx.Free()

	if frames, err = gmf.DefaultRescaler(swsCtx, frames); err != nil {
		return  errors.Wrap(err, "error rescaling")
	}

	packets, err := cc.Encode(frames, 0)
	if err != nil {
		return errors.Wrap(err, "error encoding")
	}

	if len(packets) == 0 {
		return errors.Errorf("no packets to encode\n")
	}

	for _, p := range packets {
		now := time.Now().Unix()
		if sw.lastStill == 0 {
			sw.lastStill = now
		}
		if time.Duration(now-sw.lastStill) > time.Second {
			continue
		}
		sw.lastStill = now

		sw.stills <- &Still{imgData: p.Data()}
		p.Free()
	}

	//frame, err := sw.videoDecoder.Decode(pkt.Data)
	//
	//if err != nil {
	//	return errors.Wrap(err, "error decoding packet data")
	//}
	//
	//if frame == nil {
	//	return nil
	//}
	//
	//defer frame.Free()
	//
	//// get packet time
	//if sw.lastStill == 0 {
	//	sw.lastStill = pkt.Time
	//}
	//
	//if pkt.Time-sw.lastStill < time.Duration(1*time.Second) {
	//	return nil
	//}
	//
	//sw.lastStill = pkt.Time
	//img := frame.Image
	//
	//go sw.encodeStill(&img)

	return nil
}

func (sw *StillWriter) encodeStill(img image.Image) {
	var b bytes.Buffer
	jpeg.Encode(&b, img, nil)

	sw.stills <- &Still{b.Bytes()}

}

//NewStillWriter creates a writer for passing still images through a channel for
//consumption.
func NewStillWriter(ch chan *Still) (*StillWriter, error) {
	// get video stream from streams
	//vstream, err := extractVideoStream(streams)
	//if err != nil {
	//	return nil, errors.Wrap(err, "error reading stream")
	//}
	//decoder, err := ffmpeg.NewVideoDecoder(vstream)
	//if err != nil {
	//	return nil, errors.Wrap(err, "error creating video decoder")
	//}
	//
	//if err := decoder.Setup(); err != nil {
	//	return nil, errors.Wrap(err, "error setting up video decoder")
	//}

	return &StillWriter{stills: ch}, nil
}

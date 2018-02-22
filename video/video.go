package video

import (
	"github.com/nareix/joy4/av"
	"github.com/pkg/errors"
)


func extractVideoStream(streams []av.CodecData) (av.CodecData, error) {
	for _, stream := range streams {
		if stream.Type() == av.H264 {
			return stream, nil
		}
	}

	return nil, errors.New("no h264 stream")
}

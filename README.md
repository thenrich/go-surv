# go-surv

Very basic IP camera monitoring and archiving using https://github.com/nareix/joy4 for RTSP streaming and MP4 encoding.

**Features**
- Stream from multiple cameras (RTSP/h.264)
- Access latest snapshot from each camera at http://[HOST]:[PORT]/camera/[CAMERA_NAME]
- Interval recording with option to store locally or S3

**Known issues**
- Video only, audio streams must be disabled on camera or streaming will fail

**Configuration**
```yaml
storage: s3
storageInterval: 20m
aws:
  region: us-east-1
  s3bucket: my.s3.bucket
  # Credentials must match a user with write access to s3bucket
  accessKey: [AWS_ACCESS_KEY]
  secretAccessKey: [AWS_SECRET_ACCESS_KEY]
cameras:
# Names must be unique!
- name: front_door
  source: rtsp://192.168.1.32/stream1
- name: back_door
  source: rtsp://192.168.1.34/stream1

```


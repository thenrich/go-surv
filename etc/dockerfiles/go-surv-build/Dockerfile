FROM ubuntu:18.04

RUN apt-get update && apt-get -y install software-properties-common golang-go git
RUN apt-add-repository ppa:jonathonf/ffmpeg-4
RUN apt-get update && apt-get -y install ffmpeg libswscale-dev libavcodec-dev libavformat-dev libavdevice-dev libavresample-dev


ENV PATH ${PATH}:/usr/local/go/bin


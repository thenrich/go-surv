FROM ubuntu:16.04

RUN apt-get update && apt-get -y install curl git golang-1.9-go  \
        ffmpeg libavcodec-dev libavformat-dev libavresample-dev  \
        libswscale-dev wget unzip sudo

ENV PATH ${PATH}:/usr/lib/go-1.9/bin

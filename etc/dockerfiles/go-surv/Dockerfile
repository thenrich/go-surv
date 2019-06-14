FROM ubuntu:18.04

RUN apt-get update && apt-get -y install software-properties-common golango-go git
RUN apt-add-repository ppa:jonathonf/ffmpeg-4
RUN apt-get update && apt-get -y install ffmpeg libswscale-dev libavcodec-dev libavformat-dev libavdevice-dev

#ADD bin/go-surv /usr/local/bin/go-surv

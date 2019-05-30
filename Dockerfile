FROM ubuntu:bionic 

RUN apt-get update && apt-get -y install software-properties-common golang-go git
RUN apt-add-repository ppa:jonathonf/ffmpeg-4
RUN apt-get update && apt-get -y install ffmpeg 


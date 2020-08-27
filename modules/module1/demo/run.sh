#!/usr/bin/env bash

# docker build -t ds-demo .

docker run --rm -p 8888:8888 -it  \
    --env="DISPLAY" \
    --env="QT_X11_NO_MITSHM=1" \
    --volume="/tmp/.X11-unix:/tmp/.X11-unix:rw" \
    --net=host \
    debian:stretch \
    bash
#!/bin/bash

echo "Welcome to the devcontainer!"

sudo apt update \
    && sudo apt install -y make iputils-ping bash 

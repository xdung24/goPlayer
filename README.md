# GoPlayer
A terminal based audio player

![screenshot](../assets/screenshot-ubuntu.png)

![screenshot in windows 10](../assets/screenshot-windows10.png)

## Features

* Supports mp3, flac, wav formats, ogg vorbit
* Displays metadata
* Volume controls
* Ability to rewind and fast-forward audio

## Installation

    sudo apt install libasound2-dev
    go build
    go install

This will install goPlayer to $GOPATH/bin folder.

## Usage

To open all audio files in folder: 

    goPlayer /path/to/folder/

To open one specific file: 

    goPlayer /path/to/file.mp3
    
If used without path parameter, goPlayer will assume default music folder: `~/Music/`

## Docker Build (Multi-Architecture)

To build Docker images for multiple architectures (x64, x86, arm64, armv7):

### Prerequisites

Set up QEMU emulation for ARM support (required on Intel/AMD CPUs):

```bash
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
```

Create a buildx builder instance:

```bash
docker buildx create --use --name=crossplat
```

### Building

Run the build script:

```bash
./builder.sh
```

This will:
1. Build images for x64, x86, arm64, and armv7 architectures
2. Push them to Docker Hub
3. Create a multi-architecture manifest as `dunglex/goplayer-builder:latest`

## Used libraries

* [termui](https://github.com/gizak/termui/)
* [beep](https://github.com/faiface/beep)
* [tag](https://github.com/dhowden/tag/)
* [oggvorbis](https://github.com/jfreymuth/oggvorbis/)
* 
## License
MIT

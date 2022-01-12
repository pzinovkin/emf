# Emf

Enhanced Metafile Format ([MS-EMF](https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-emf/)) reader and converter. Supports only png output.

## Installation

    go get github.com/pzinovkin/emf/cmd/emftopng

## Usage

Calling by passing file path as an argument will generate image in the folder next to source emf file.

    emftopng /path/to/image.emf

Also supports stdin. Image will be written to stdout.

    emftopng < image.emf > image.png

# emftoimg

Converts Enhanced Metafile Format files ([MS-EMF](http://msdn.microsoft.com/en-us/library/cc230514.aspx)) to other image formats.

Supports only png output for now.

## Installation

    go get github.com/pzinovkin/emftoimg

## Usage

Calling by passing file path as argument will generate image in the folder next to source emf file.

    $ emftoimg /path/to/image.emf

Also supports stdin. Image will be writted to stdout.

    $ cat /path/to/image.emf | emftoimg > /new/path/image.png



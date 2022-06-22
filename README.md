# Google Image Upsizer
[![imageupsizer](https://github.com/kmulvey/imageupsizer/actions/workflows/release_build.yml/badge.svg)](https://github.com/kmulvey/imageupsizer/actions/workflows/release_build.yml)
Extract the best images from Google Image Search. Divergent fork of [yms2772/google_image_upsizer]([yms2772/google_image_upsizer](https://github.com/yms2772/google_image_upsizer)).

## How to use
```go
go install github.com/kmulvey/imageupsizer
```
```go
imageupsizer -path="original/test.jpg" -output="result/"
```

## Required options
|Name|Type|Description|
|------|---|---|
|`-input`|`string`|Path to the image file or directory you want to upscale|
|`-output`|`string`|Directory path to save results|

# Result
![test](https://user-images.githubusercontent.com/6222645/167277591-7f92d665-7e92-4698-8d0a-216d44170c3d.png)
![test2](https://user-images.githubusercontent.com/6222645/167277593-61beab00-259b-4ebe-bb79-60dd4b4d084b.png)

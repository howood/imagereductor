[![Build Status](https://travis-ci.org/howood/imagereductor.svg?branch=master)](https://travis-ci.org/howood/imagereductor)
[![GitHub release](http://img.shields.io/github/release/howood/imagereductor.svg?style=flat-square)][release]
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/howood/imagereductor)
[![Test Coverage](https://api.codeclimate.com/v1/badges/00e0b66cf675d519a2a8/test_coverage)](https://codeclimate.com/github/howood/imagereductor/test_coverage)
[![Go Report Card](https://goreportcard.com/badge/github.com/howood/imagereductor)](https://goreportcard.com/report/github.com/howood/imagereductor)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]

[release]: https://github.com/howood/imagereductor/releases
[license]: https://github.com/howood/imagereductor/blob/master/LICENSE

# imagereductor

Image delivery from storage of AWS S3 / GCS contents with Resizing and Caching

## URL Query Key to Resize

* key : path of storage
* w : 500 (px)  | risizeing width with original aspect ratio
* h : 500 (px) | risizeing height with original aspect ratio
* q : 1 ~ 4      | change image quality
* rotate : right,left,upsidedown,autovertical,autohorizontal  | auto... are rotate image auto vertical position / horizontal position
* crop : 111,222,333,444 (from point x/y - to point x/y)
* bri : 0 ~ 100   | change image brightness
* cont : -100 ~ 100   | change image contrast
* gam : 0.0 ~     | change image gamma
* nonusecache: true

## Form Key to Upload(multipart/form-data)

* path : path of storage
* uploadfile : filepath

## Endpoint
| Method        | endpoint          | usage          |
| --------------- |---------------|---------------|
| GET | / | Get image file using query options |
| POST | / | Upload image file with bearer token of authorization header|
| GET | /files | Get non-image file using 'key' query option only |
| POST | /files | Upload non-image file with bearer token of authorization header|
| GET | /streaming | Get non-image file using 'key' query option only with HTTP Streaming |
| GET | /info | Get file (Content-Type / Content-Length) info using 'key' and 'nonusecache' query option only |
| GET | /token | Get bearer token (Only IP addresses restricted by TOKENAPI_ALLOW_IPS can be requested) |

## using docker

| env        | param          |
| --------------- |---------------|
| VERIFY_MODE |enable / disable |
| LOG_MODE |minimum / few or empty |
| ADMIN_MODE |enable / disable |
| SERVER_PORT |8080, 80, etc |
| TOKENAPI_ALLOW_IPS |72.22.0.1/24,127.0.0.1/32(separate with comma) |
| CACHE_TYPE |redis / gocache |
| REDISHOST |x.x.x.x |
| REDISPORT |6379 |
| REDISTLS |use or empty |
| REDISPASSWORD | |
| CACHEDDB |0~ |
| CACHEEXPIED |300 (seconds) |
| HEADEREXPIRED |300 (seconds) |
| STORAGE_TYPE |s3 / gcs |
| AWS_S3_LOCALUSE |use or empty (use with minio) |
| AWS_S3_REGION | |
| AWS_S3_BUKET | |
| AWS_S3_ACCESSKEY | |
| AWS_S3_SECRETKEY | |
| AWS_S3_ENDPOINT |(use with minio) |
| GCS_BUKET | |
| GCS_PROJECTID | |
| GOOGLE_APPLICATION_CREDENTIALS | |
| TOKEN_SECRET |(use with jwt token when upload images) |
| VALIDATE_IMAGE_TYPE | jpeg,gif,png,bmp,tiff |
| VALIDATE_IMAGE_MAXWIDTH |5000 (px) |
| VALIDATE_IMAGE_MAXHEIGHT |5000 (px) |
| VALIDATE_IMAGE_MAXFILESIZE |104857600 (byte) |

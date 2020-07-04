# imagereductor

Image delivery from storage of AWS S3 / GCS contents with Resizing and Caching

## environment parameters

| env        | param          |
| --------------- |---------------|
| VERIFY_MODE |enable / disable |
| LOG_MODE |minimum / few or empty |
| ADMIN_MODE |enable / disable |
| SERVER_PORT |8080, 80, etc |
| CACHE_TYPE |redis / gocache |
| REDISHOST |x.x.x.x |
| REDISPORT |6379 |
| REDISTLS |use or empty |
| REDISPASSWORD | |
| CACHEDDB |0~ |
| CACHEEXPIED |30 (seconds) |
| STORAGE_TYPE |s3 / gcs |
| AWS_S3_LOCALUSE |use or empty (use with minio) |
| AWS_S3_REGION | |
| AWS_S3_BUKET | |
| AWS_S3_ACCESSKEY | |
| AWS_S3_SECRETKEY | |
| AWS_S3_ENDPOINT |(use with minio |
| GCS_BUKET | |
| GCS_PROJECTID | |
| GOOGLE_APPLICATION_CREDENTIALS | |
| TOKEN_SECRET |(use with jwt token when upload images) |
| VALIDATE_IMAGE_TYPE | jpeg,gif,png,bmp,tiff |
| VALIDATE_IMAGE_MAXWIDTH |5000 (px) |
| VALIDATE_IMAGE_MAXHEIGHT |5000 (px) |
| VALIDATE_IMAGE_MAXFILESIZE |5000000 (byte) |

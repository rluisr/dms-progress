dms-progress
=============
Get AWS DMS replication progress and send it to slack running on Lambda.

![](https://github.com/rluisr/image-store/blob/master/dms-progress/dms-progress01.png?raw=true)

Usage
-----
1. Clone
2. Build
3. Upload
```
go build -a -tags netgo -installsuffix netgo --ldflags '-extldflags "-static"' -o main && zip main.zip ./main && mv -f main.zip ~/Desktop/ && rm -rf main
```

handler is `main` and you have to set these environment variables.
- `SLACK_INCOMING_URL`
- `REGION`

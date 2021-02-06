# lodestone-publisher


# How to build

```bash
docker build -f Dockerfile.fs -t lodestone-fs-publisher .
docker run lodestone-fs-publisher

docker build -f Dockerfile.email -t lodestone-email-publisher .
docker run lodestone-email-processor
```


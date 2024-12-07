# image-mirror

On-demand mirror for images from dockerhub.

Check if is [there](https://github.com/orgs/1mgr/packages?repo_name=image-mirror) a mirror for the image you want to use.

Or just run the following command:

```bash
# put your image name and tag, for example alpine:3.21.0
curl -NL 1mgr.xyz/<image:tag>
# or
wget -qO- 1mgr.xyz/<image:tag>
```

## Usage demo

![render1733588375570](https://github.com/user-attachments/assets/a8366fd0-5bee-4536-8b52-31ed053eb309)

## Why?

Since dockerhub is getting slow and unreliable with their [rate limits](https://docs.docker.com/docker-hub/download-rate-limit/) it's not a good idea to rely on it for production deployments. This project aims to provide an on-demand mirror for needed images.

## Disclaimer

All images are mirrored from dockerhub as is, we do not own any of them.

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

This project was created as a personal tool and is not intended for production use. While it may be useful for self-hosters and individuals, please note:

* The service is public-facing and may be prone to abuse (e.g., DDOS, resource exhaustion).
* It relies on GitHub Actions and Packages, which are free for open-source projects but subject to limits.
* Only public Docker images are used, with no authentication. Abuse could impact GitHub Action runner IPs.
* Security measures are in place, but risks remain due to the open-source nature of the project.
* All images are mirrored from dockerhub as is, we do not own any of them.

Use responsibly and implement additional safeguards if needed. Feedback and contributions are welcome!

## Limits

### RPS

There is a limit of 10 requests per 10 seconds per IP.
After exceeding the limit, the service will return a 429 status code with message `error code: 1015`.

### Mirror of existing images

The service will only mirror images that are not mirrored yet or the mirror is older than 12 hours.

# Noob Store

Trying to build a s3 clone from scratch.
Written in Golang, this is a mono repo that holds the api as well as the fs layer.

Here's my tweet showcasing the full product

<blockquote class="twitter-tweet" data-media-max-width="560"><p lang="en" dir="ltr">This is NoobStore, my blob storage, handling 50 requests coming from 8 client like a pro!<br>Avg: 1 second! <a href="https://t.co/SyipRF99Kf">pic.twitter.com/SyipRF99Kf</a></p>&mdash; NoobScience (@NoobScience1) <a href="https://twitter.com/NoobScience1/status/1880676341496123843?ref_src=twsrc%5Etfw">January 18, 2025</a></blockquote> 

## Implemented 

- [x] FS Layer
- [x] InsertBlob
- [x] GetBlob
- [x] Metadata Layer
- [x] Handler
- [x] Api Layer
- [x] Authentication
- [x] Horizontal Scaling (using Docker)
- [x] Add Cache layer
- [x] Logs
- [x] Atomic Metadata Operations
- [x] Optimize Blob storage
- [x] Bucket Delegation Algorithm
- [x] Secondary Writes to Bucket (Free up space)
- [x] Automatic Buckets Expand
- [x] Fix Concurrency Issues
- [ ] Atomic FS Layer Operations

## License

The above project is licensed under the MIT License. More details can be found in the [LICENSE](LICENSE) file.

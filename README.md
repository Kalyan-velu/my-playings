## MY Playings
### Summary

---
This is my playground to learn Golang without any help of AI tools to generate code.
Goal is to learn Go language. Going deeper into golang\'s features.
---

### What I learned

---
### Golang
- Using my learnings of Go to build a web server to serve my music playlists from Spotify and Youtube .
- Learned how to write unit tests for my application.
- Learned how to use Go's built-in testing framework.
- Learned how Go's concurrency works.
- Learned how to use Go's modules.
- Learned how to handle oauth2.0 authentication with Google and Spotify Services.
- Learned how context works in Go.
- Learned how `stuct` work in Go.
- Learned how to use `sync` package to work with concurrency.
---
### Docker
- Learned how to dockerize my application.
- Learned how to create a CI/CD pipeline to build an image of my application and deploy it to render.
- Learned how to debug my application using Remote Debugging tools.
- Learned how to encrypt file contents using AES.
---
### Other
- Learned how to use `make` to automate repetitive tasks.
- Learned how to use `air` with the debugger to debug my application.
- Learned how to use `goland` to debug my application.

## Problems I faced (and how I solved them)

---
### Golang

**Structs**: I learned how to model data by composing structs and using pointers when I needed shared state. Defining methods with pointer receivers (e.g. `func (s *Song) Play() {}`) helped me understand mutation and method sets.

**Concurrency**: I practiced protecting shared state with `sync.Mutex` and read/write patterns with `sync.RWMutex`. I now lock around critical sections and keep the lock scope small.

**net/http**: I switched from the default `http.HandleFunc` to a `ServeMux` to keep routing and middleware clean. I learned how path parameters flow into handlers via `r.PathValue("param")` and how to set timeouts on custom `http.Client`s for predictable behavior. I also built a simple logging middleware and standardized error handling.

**oauth2**: I separated provider setup (`oauth2.Config`) from token usage (`TokenSource`) and learned how to refresh and persist tokens. For Google/YouTube and Spotify, I now use the same token-store flow and request clients per call, which keeps providers consistent.

**debugging**:
- With `air`, I run the debugger against the reloaded binary and keep logs clean so I can spot reload errors quickly.
- For remote debugging, I expose the debug port and attach from my IDE.
- In GoLand, I use run configurations and attach to the process when needed.

**testing**:
- I started with small unit tests around provider wiring and error cases, then expanded coverage as features stabilized.

**modules**:
- I learned how `go mod` resolves public vs private modules, how to configure authentication for private repos, and how to structure packages so modules are reusable across folders.

### Docker
- I learned to containerize the app with a multi-stage build for smaller images, push to a registry, and deploy on [render](https://render.com).
- Faced problems with writing encrypted files to disk.
- Learned how to use `docker-compose` for local development and testing.

### CI/CD
- I automated builds and deployments by creating a pipeline that builds the image and ships it to render.
---
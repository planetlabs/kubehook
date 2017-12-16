# kubehook  [![Godoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/negz/kubehook) [![Travis](https://img.shields.io/travis/negz/kubehook.svg?maxAge=300)](https://travis-ci.org/negz/kubehook/) [![Codecov](https://img.shields.io/codecov/c/github/negz/kubehook.svg?maxAge=3600)](https://codecov.io/gh/negz/kubehook/)
Kubehook is a [webhook token authentication](https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication)
service for Kubernetes. It provides one API endpoint to generate
[JSON Web Tokens](https://jwt.io), and another to validate tokens on behalf of
Kubernetes. Kubehook assumes it is running behind a reverse proxy that
terminates TLS and authenticates callers.

## Generating a token
Kubehook provides a small web UI to request tokens:
![Request a token](frontend/gentoken.png)

Once a token has been requested the UI explains how to use it:
![Configure a token](frontend/usetoken.png)

## Usage
Tokens will be generated for the username provided via a configurable HTTP
header - `X-Forwarded-User` by default.

To generate a token with a 24 hour lifetime:
```bash
$ export USERNAME=cooluser
$ curl -i -X POST \
	-H "Content-Type: application/json" \
	-H "X-Forwarded-User: ${USERNAME}" \
	-d "{\"lifetime\": \"24h\"}" \  # Lifetime is parsed via https://golang.org/pkg/time/#ParseDuration
	http://localhost:10003/generate
  
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 11 Dec 2017 08:00:14 GMT
Content-Length: 208

{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJnaXRodWIuY29tL25lZ3ova3ViZWhvb2siLCJleHAiOjE1MTMwNjU2MTQsIm5iZiI6MTUxMjk3OTIxNCwic3ViIjoiY29vbHVzZXIifQ.ylxC6wFglLBI0_0e0LQftwuqI95mb0L5Wjnp5aCmPHU"}
```

To validate a token:
```bash
export TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJnaXRodWIuY29tL25lZ3ova3ViZWhvb2siLCJleHAiOjE1MTMwNjU2MTQsIm5iZiI6MTUxMjk3OTIxNCwic3ViIjoiY29vbHVzZXIifQ.ylxC6wFglLBI0_0e0LQftwuqI95mb0L5Wjnp5aCmPHU
curl -i -X GET \
    -H "Content-Type: application/json" \
    -d "{\"apiVersion\":\"authentication.k8s.io/v1beta1\",\"kind\":\"TokenReview\",\"spec\":{\"token\":\"${TOKEN}\"}}" \
    http://localhost:10003/authenticate

HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 11 Dec 2017 08:02:10 GMT
Content-Length: 239

{"kind":"TokenReview","apiVersion":"authentication.k8s.io/v1beta1","metadata":{"creationTimestamp":"2017-12-11T08:02:10Z"},"spec":{},"status":{"authenticated":true,"user":{"username":"cooluser","uid":"github.com/negz/kubehook/cooluser"}}}
```

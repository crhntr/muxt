You can write your own version of the execute function.

Just copy the signature and replace the body as you like.

Remember to honor the "writeHeader" parameter.
If a method receives a response writer, you should expect that method call WriteHeader and not do so in your execute implementation.

Your function has to have the following signature:

```go
func execute(response http.ResponseWriter, request *http.Request, writeHeader bool, name string, code int, data any) {}
```
# assign

[![Go Reference](https://pkg.go.dev/badge/github.com/norunners/assign/tree/dev.svg)](https://pkg.go.dev/github.com/norunners/assign/tree/dev)

Package `assign` assigns values of any source to Go values.

### Examples

Assign to a Go value from any source.
```go
err := assign.ToFrom(dst, src)
```

Assign from any source to a Go value.
```go
err := assign.From(src).To(dst)
```

Assign with assign.WithTags option.
```go
err := assign.ToFrom(dst, src, assign.WithTags("json"))
```

Assign with assign.WithoutCycle option.
```go
err := assign.ToFrom(dst, src, assign.WithoutCycle())
```

Assign from an assign.Source.
```go
src := assign.Of(val)
err := assign.ToFrom(dst, src)
```

Assign from assign.Assigner to multiple Go values.
```go
assigner := assign.From(src)
err := assigner.To(dst)
errAnother := assigner.To(dstAnother)
```

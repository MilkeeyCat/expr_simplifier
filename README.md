# expr_simplifier
Simple maffs expression simplifier(simplificator?) using e-graphs.

In case you don't know what e-graphs are, checkout:
- https://www.cole-k.com/2023/07/24/e-graphs-primer
- https://en.wikipedia.org/wiki/E-graph

## Example of usage
To use this puppy you need 2 things:
- file with rewrite rules(in a format `pattern => rewrite`, for example `a + 0
=> a`), each rule has to be on a separate line.
- you need only the file really.

Run `go run ./cmd/main.go -rules ./your/path/to/rewrite.rules "the expression you want to simplify"`.

<img src="/images/example.png"/>

There's also a way to produce a resulting e-graph in SVG format, to do that,
pass `-viz` flag.

## Note on produced visualization
It's not quite correct. The nodes that have edges to other nodes in the same
cluster(grey box) actually point the the box itself, but graphviz doesn't allow
such edges.
So
<img src="/images/bad_graph.png"/>
is actually
<img src="/images/good_graph.png"/>

package egraph

import (
	"context"
	"io"
	"strconv"

	"github.com/goccy/go-graphviz"
)

func Visualize(ctx context.Context, egraph *Egraph, dest io.Writer) error {
	g, err := graphviz.New(ctx)
	if err != nil {
		return err
	}

	graph, err := g.Graph()
	if err != nil {
		return err
	}

	graph = graph.SetCompound(true)

	collector := new(collector)
	classToName := make(map[*eclass]string)
	nodeToName := make(map[enode]string)

	egraph.DFS(collector)

	for classIdx, class := range collector.classes {
		clusterName := "cluster_" + strconv.Itoa(classIdx)
		cluster, err := graph.CreateSubGraphByName(clusterName)
		if err != nil {
			return err
		}

		classToName[class] = clusterName
		cluster = cluster.SetStyle(graphviz.FilledGraphStyle)

		for nodeIdx, node := range class.nodes {
			nodeName := "node_" + strconv.Itoa(classIdx) + "_" + strconv.Itoa(nodeIdx)
			graphNode, err := cluster.CreateNodeByName(nodeName)
			if err != nil {
				return err
			}

			nodeToName[node] = nodeName
			graphNode = graphNode.SetLabel(node.String(egraph))
			graphNode = graphNode.SetShape(graphviz.RectShape)
		}
	}

	for i, edge := range collector.edges {
		tail, err := graph.NodeByName(nodeToName[edge.node])
		if err != nil {
			return err
		}

		cluster, err := graph.SubGraphByName(classToName[edge.class])
		if err != nil {
			return err
		}

		head, err := cluster.FirstNode()
		if err != nil {
			return err
		}

		graphEdge, err := graph.CreateEdgeByName("edge_"+strconv.Itoa(i), tail, head)
		if err != nil {
			return err
		}

		graphEdge.SetLogicalHead(classToName[edge.class])
	}

	if err := g.Render(ctx, graph, graphviz.SVG, dest); err != nil {
		return err
	}

	if err := graph.Close(); err != nil {
		return err
	}

	if err := g.Close(); err != nil {
		return err
	}

	return nil
}

type edge struct {
	node  enode
	class *eclass
}

type collector struct {
	classes []*eclass
	edges   []edge
}

func (c *collector) VisitEclass(class *eclass) {
	c.classes = append(c.classes, class)

	for _, node := range class.nodes {
		for _, class := range node.Children() {
			c.edges = append(c.edges, edge{
				node:  node,
				class: class,
			})
		}
	}
}

func (c *collector) VisitEnode(node enode) {}

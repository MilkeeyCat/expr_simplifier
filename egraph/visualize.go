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
	classIDToName := make(map[EclassID]string)
	nodeToName := make(map[Enode]string)

	egraph.DFS(collector)

	for classIdx, classID := range collector.classIDs {
		clusterName := "cluster_" + strconv.Itoa(classIdx)
		cluster, err := graph.CreateSubGraphByName(clusterName)
		if err != nil {
			return err
		}

		classIDToName[classID] = clusterName
		cluster = cluster.SetStyle(graphviz.FilledGraphStyle)

		for nodeIdx, node := range egraph.Eclass(classID).nodes {
			nodeName := "node_" + strconv.Itoa(classIdx) + "_" + strconv.Itoa(nodeIdx)
			graphNode, err := cluster.CreateNodeByName(nodeName)
			if err != nil {
				return err
			}

			nodeToName[node] = nodeName
			graphNode.SetLabel(node.String(egraph))
			graphNode.SetShape(graphviz.RectShape)
		}
	}

	for i, edge := range collector.edges {
		tail, err := graph.NodeByName(nodeToName[edge.node])
		if err != nil {
			return err
		}

		cluster, err := graph.SubGraphByName(classIDToName[edge.classID])
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

		graphEdge.SetLogicalHead(classIDToName[edge.classID])
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
	node    Enode
	classID EclassID
}

type collector struct {
	classIDs []EclassID
	edges    []edge
}

func (c *collector) VisitEclass(graph *Egraph, classID EclassID) {
	c.classIDs = append(c.classIDs, classID)

	for _, node := range graph.Eclass(classID).nodes {
		for _, classID := range node.Children() {
			c.edges = append(c.edges, edge{
				node:    node,
				classID: classID,
			})
		}
	}
}

func (c *collector) VisitEnode(graph *Egraph, node Enode) {}

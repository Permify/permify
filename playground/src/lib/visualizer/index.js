import React, {useEffect, useRef, useState} from 'react';
import {DataSet, Network} from 'vis-network/standalone';
import GraphOptions from "./config";

function Visualizer(props) {
    const containerRef = useRef(null);
    const networkRef = useRef(null);

    // data
    const [graph, setGraph] = useState({nodes: [], edges: []});

    function modifyNodes() {
        return new Promise((resolve) => {
            let nodes = []
            for (const index in props.graph.nodes) {
                if (props.graph.nodes[index].type === "operation") {

                    let label = ""

                    if (props.graph.nodes[index].label === "OPERATION_UNION") {
                        label = "or"
                    }

                    if (props.graph.nodes[index].label === "OPERATION_INTERSECTION") {
                        label = "and"
                    }

                    if (props.graph.nodes[index].label === "OPERATION_EXCLUSION") {
                        label = "not"
                    }

                    nodes.push({
                        id: props.graph.nodes[index].id,
                        label: label,
                        group: props.graph.nodes[index].type
                    })
                } else {
                    nodes.push({
                        id: props.graph.nodes[index].id,
                        label: props.graph.nodes[index].label,
                        group: props.graph.nodes[index].type
                    })
                }
            }
            resolve(nodes);
        });
    }

    function modifyEdges() {
        return new Promise((resolve) => {
            let edges = []
            for (const index in props.graph.edges) {
                switch (props.graph.edges[index].from.type) {
                    case "entity":
                        edges.push({
                            from: props.graph.edges[index].from.id,
                            to: props.graph.edges[index].to.id,
                            color: {color: 'rgba(99,24,255,0.4)', inherit: false},
                            dashes: false,
                            arrows: {
                                to: {
                                    enabled: false
                                }
                            }
                        })
                        break
                    case "rule":
                        edges.push({
                            from: props.graph.edges[index].from.id,
                            to: props.graph.edges[index].to.id,
                            color: {color: 'rgba(99,24,255,0.4)', inherit: false},
                            dashes: false,
                            arrows: {
                                to: {
                                    enabled: false
                                }
                            }
                        })
                        break
                    case "relation":
                        edges.push({
                            from: props.graph.edges[index].from.id,
                            to: props.graph.edges[index].to.id,
                            color: {color: 'rgba(147,241,238,0.4)', inherit: false},
                            dashes: false
                        })
                        break
                    case "attribute":
                        edges.push({
                            from: props.graph.edges[index].from.id,
                            to: props.graph.edges[index].to.id,
                            color: {color: 'rgba(255,165,0,0.4)', inherit: false},
                            dashes: false,
                            arrows: {
                                to: {
                                    enabled: false
                                }
                            }
                        })
                        break
                    case "permission":
                        edges.push({
                            from: props.graph.edges[index].from.id,
                            to: props.graph.edges[index].to.id,
                            color: {color: 'rgba(91,204,99,0.4)', inherit: false},
                            dashes: false
                        })
                        break
                    case "operation":
                        if (props.graph.edges[index].from.label === "OPERATION_UNION") {
                            edges.push({
                                from: props.graph.edges[index].from.id,
                                to: props.graph.edges[index].to.id,
                                color: {color: 'rgba(229,52,114,0.4)', inherit: false},
                                dashes: true,
                            })
                        }

                        if (props.graph.edges[index].from.label === "OPERATION_INTERSECTION") {
                            edges.push({
                                from: props.graph.edges[index].from.id,
                                to: props.graph.edges[index].to.id,
                                color: {color: 'rgba(229,52,114,0.4)', inherit: false},
                                dashes: false,
                            })
                        }

                        if (props.graph.edges[index].from.label === "OPERATION_EXCLUSION") {
                            edges.push({
                                from: props.graph.edges[index].from.id,
                                to: props.graph.edges[index].to.id,
                                color: {color: 'rgba(229,52,114,0.4)', inherit: false},
                                dashes: false,
                            })
                        }
                        break
                    default:
                        edges.push({
                            from: props.graph.edges[index].from.id,
                            to: props.graph.edges[index].to.id,
                        })
                }
            }
            resolve(edges);
        });
    }

    function makeGraph() {
        modifyNodes().then(n => {
            return n
        }).then(nn => {
            modifyEdges().then(e => {
                return {nodes: nn, edges: e}
            }).then(r => {
                if (r.nodes.length > 0) {
                    setGraph(r)
                }
            })
        })
    }

    useEffect(() => {
        setGraph({nodes: [], edges: []})
        if (props.graph) {
            makeGraph()
        }
    }, [props.graph]);

    useEffect(() => {
        if (!containerRef.current || graph.nodes.length === 0) {
            networkRef.current?.destroy();
            networkRef.current = null;
            return;
        }

        const network = new Network(
            containerRef.current,
            {
                nodes: new DataSet(graph.nodes),
                edges: new DataSet(graph.edges),
            },
            GraphOptions()
        );

        networkRef.current = network;

        return () => {
            network.destroy();
            networkRef.current = null;
        };
    }, [graph]);

    return (
        <div style={{height: "100vh"}}>
            <div ref={containerRef} style={{height: "100%"}} />
        </div>
    );
}

export default Visualizer;

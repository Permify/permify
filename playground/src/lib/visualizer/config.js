function GraphOptions() {
    return  {
        autoResize: true,
        clickToUse: true,
        height: '90%',
        width: '100%',
        layout: {
            hierarchical: {
                enabled: true,
                direction: "UD",
                sortMethod: "directed",
                shakeTowards: "roots",
                levelSeparation: 150,
                nodeSpacing: 220,
                treeSpacing: 150,
                blockShifting: true,
                edgeMinimization: true,
                parentCentralization: true,
            }
        },
        interaction: {
            tooltipDelay: 10000,
            navigationButtons: true,
            keyboard: false,
            hover: true,
            multiselect: true,
            hoverConnectedEdges: true
        },
        physics: {
            enabled: false,
        },
        nodes: {
            fixed: {
                x: false,
                y: false
            },
            color: {
                hover: {
                    border: "#8246FF",
                    background: "#8246FF",
                }
            },
            font: {
                color: '#ffffff',
                size: 20
            },
            shape: "dot",
            size: 25,
            scaling: {
                min: 10,
                max: 60,
                label: {
                    enabled: true,
                    min: 20,
                    max: 32
                }
            }
        },
        edges: {
            hoverWidth: 1,
            arrows: {
                to: {
                    enabled: true,
                    scaleFactor: 1,
                    type: "arrow"
                }
            },
            color: {
                color: "#8246FF",
                highlight: "#8246FF",
                hover: "#8246FF",
                inherit: 'from'
            },
            font: {
                size: 20,
                color: "white",
                strokeWidth: 6,
                strokeColor: "#141517"
            },
            width: 3,
            smooth: true
        },
        groups: {
            entity: {
                color: { background: "#6318FF", border: "#6318FF" },
                scaling: { min: 20 },
                shape: "dot",
                size: 30,
            },
            rule: {
                color: { background: "#f718ff", border: "#cd18ff" },
                scaling: { min: 20 },
                shape: "square",
                size: 20,
            },
            relation: {
                color: { background: "#93F1EE", border: "#93F1EE" },
                scaling: { min: 10 },
                shape: "dot",
                size: 20,
            },
            attribute: {
                color: { background: "#FFA500", border: "#FFA500" },
                scaling: { min: 10 },
                shape: "dot",
                size: 20,
            },
            permission: {
                color: { background: "#5bcc63", border: "#5bcc63" },
                scaling: { min: 10 },
                shape: "dot",
                size: 20,
            },
            operation: {
                color: { background: "#e53472", border: "#e53472" },
                shapeProperties: { borderDashes: false },
                shape: "icon",
                icon: {
                    face: "FontAwesome",
                    code: "\uf286",
                    size: 40,
                    color: "#e53472"
                },
                size: 10,
            },
        },
    };
}

export default GraphOptions
